package security

import (
	"errors"
	"testing"
)

func TestLocalEncryptedVault_SudoPasswordPerHost(t *testing.T) {
	v := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))

	// No sudo password configured yet -> must error (Ansible would prompt / fail).
	if _, err := v.GetSudoPassword("web-prod-01"); err == nil {
		t.Fatal("expected error when sudo password is not configured")
	}

	// Store a per-host sudo password, encrypted at rest.
	if err := v.EncryptAndStoreSudoPassword("web-prod-01", "s3cr3t-sudo"); err != nil {
		t.Fatalf("failed to store sudo password: %v", err)
	}

	got, err := v.GetSudoPassword("web-prod-01")
	if err != nil {
		t.Fatalf("unexpected error retrieving sudo password: %v", err)
	}
	if got != "s3cr3t-sudo" {
		t.Fatalf("expected per-host sudo password 's3cr3t-sudo', got %q", got)
	}

	// Different host must not leak another host's secret.
	if _, err := v.GetSudoPassword("db-master"); !errors.Is(err, ErrSudoPasswordMissing) {
		t.Fatalf("expected ErrSudoPasswordMissing for unrelated host, got %v", err)
	}
}

func TestLocalEncryptedVault_SudoPasswordSeparateFromKey(t *testing.T) {
	v := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))

	_ = v.EncryptAndStore("web-prod-01", "private-key-bytes")
	_ = v.EncryptAndStoreSudoPassword("web-prod-01", "s3cr3t-sudo")

	key, err := v.GetPrivateKey("web-prod-01")
	if err != nil || key != "private-key-bytes" {
		t.Fatalf("private key retrieval broken: %q %v", key, err)
	}
	pass, err := v.GetSudoPassword("web-prod-01")
	if err != nil || pass != "s3cr3t-sudo" {
		t.Fatalf("sudo password retrieval broken: %q %v", pass, err)
	}
}

func TestApplySudoPasswordEnv(t *testing.T) {
	v := NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))

	environ := []string{
		"SUDO_PASS_WEB_PROD_01=s3cr3t-web",
		"SUDO_PASS_DB_MASTER=s3cr3t-db",
		"OPENAI_API_KEY=should-be-ignored",
		"PATH=/usr/bin",
	}

	ApplySudoPasswordEnv(v, environ)

	got, err := v.GetSudoPassword("web-prod-01")
	if err != nil || got != "s3cr3t-web" {
		t.Fatalf("expected web-prod-01 sudo password 's3cr3t-web', got %q %v", got, err)
	}
	got, err = v.GetSudoPassword("db-master")
	if err != nil || got != "s3cr3t-db" {
		t.Fatalf("expected db-master sudo password 's3cr3t-db', got %q %v", got, err)
	}
	// Unrelated hosts remain unset.
	if _, err := v.GetSudoPassword("cache-01"); !errors.Is(err, ErrSudoPasswordMissing) {
		t.Fatalf("expected cache-01 to be unset, got %v", err)
	}
}
