package security

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalEncryptedVault_APIKeyRoundTrip(t *testing.T) {
	v := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))

	// No API key configured yet -> must error.
	if _, err := v.GetAPIKey(); err == nil {
		t.Fatal("expected error when API key is not configured")
	}

	if err := v.StoreAPIKey("sk-or-abc123"); err != nil {
		t.Fatalf("failed to store API key: %v", err)
	}

	got, err := v.GetAPIKey()
	if err != nil {
		t.Fatalf("unexpected error retrieving API key: %v", err)
	}
	if got != "sk-or-abc123" {
		t.Fatalf("expected API key 'sk-or-abc123', got %q", got)
	}

	// API key must not collide with per-host secrets stored under host aliases.
	if _, err := v.GetPrivateKey("web-prod-01"); err == nil {
		t.Fatal("API key leaked into host alias namespace")
	}
}

func TestLocalEncryptedVault_SaveLoadPersistsEncrypted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vault.enc")

	v1 := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))
	if err := v1.StoreAPIKey("sk-or-abc123"); err != nil {
		t.Fatalf("store api key: %v", err)
	}
	if err := v1.EncryptAndStoreSudoPassword("web-prod-01", "s3cr3t-sudo"); err != nil {
		t.Fatalf("store sudo: %v", err)
	}
	if err := v1.Save(path); err != nil {
		t.Fatalf("save vault: %v", err)
	}

	// A fresh vault loaded from disk must recover both secrets.
	v2 := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))
	if err := v2.Load(path); err != nil {
		t.Fatalf("load vault: %v", err)
	}

	got, err := v2.GetAPIKey()
	if err != nil || got != "sk-or-abc123" {
		t.Fatalf("expected API key 'sk-or-abc123' after load, got %q %v", got, err)
	}
	sudo, err := v2.GetSudoPassword("web-prod-01")
	if err != nil || sudo != "s3cr3t-sudo" {
		t.Fatalf("expected sudo 's3cr3t-sudo' after load, got %q %v", sudo, err)
	}
}

func TestLocalEncryptedVault_LoadMissingFileIsNoOp(t *testing.T) {
	v := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))
	// Loading a non-existent vault file must not error (first-run bootstrap).
	if err := v.Load(filepath.Join(t.TempDir(), "does-not-exist.enc")); err != nil {
		t.Fatalf("expected nil error for missing vault file, got %v", err)
	}
	if _, err := v.GetAPIKey(); !errors.Is(err, ErrAPIKeyMissing) {
		t.Fatalf("expected ErrAPIKeyMissing after empty load, got %v", err)
	}
}

func TestApplyAPIKeyEnv(t *testing.T) {
	v := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))

	environ := []string{
		"OPENAI_API_KEY=sk-or-from-env",
		"SUDO_PASS_WEB_PROD_01=s3cr3t-web",
		"PATH=/usr/bin",
	}

	ApplyAPIKeyEnv(v, environ)

	got, err := v.GetAPIKey()
	if err != nil || got != "sk-or-from-env" {
		t.Fatalf("expected API key 'sk-or-from-env', got %q %v", got, err)
	}

	// Unrelated env vars must be ignored.
	if _, err := v.GetPrivateKey("web-prod-01"); err == nil {
		t.Fatal("ApplyAPIKeyEnv must not seed host aliases")
	}
	_ = os.Setenv("OPENAI_API_KEY", "sk-or-from-env")
}
