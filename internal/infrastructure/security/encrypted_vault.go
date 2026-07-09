package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/devops/agent/internal/domain/agent"
)

// ErrSudoPasswordMissing indicates no per-host sudo password is configured in
// the vault. Callers should fall back to passwordless sudo (sudo -n).
var ErrSudoPasswordMissing = errors.New("sudo password missing for target alias")

// ErrAPIKeyMissing indicates no LLM API key is configured in the vault.
var ErrAPIKeyMissing = errors.New("openai api key missing from vault")

// apiKeyAlias namespaces the LLM API key under a reserved key so it never
// collides with per-host secrets stored under host aliases.
const apiKeyAlias = "llm:openai"

// vaultFileName is the on-disk encrypted vault (mirrors ansible-vault's single
// encrypted blob; here it holds the already-encrypted KeyStore map).
const vaultFileName = "vault.enc"

// sudoKeySuffix namespaces the sudo password under the host alias so it never
// collides with the host's private key (mirrors Ansible keeping
// ansible_ssh_private_key_file and ansible_become_password distinct).
const sudoKeySuffix = ":sudo"

type LocalEncryptedVault struct {
	MasterKey []byte
	KeyStore  map[string]string 
}

func NewLocalEncryptedVault(masterKey []byte) *LocalEncryptedVault {
	return &LocalEncryptedVault{
		MasterKey: masterKey,
		KeyStore:  make(map[string]string),
	}
}

func (v *LocalEncryptedVault) EncryptAndStore(hostAlias, plaintext string) error {
	block, err := aes.NewCipher(v.MasterKey)
	if err != nil {
		return fmt.Errorf("context: %w", err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("context: %w", err)
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("context: %w", err)
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	v.KeyStore[hostAlias] = base64.StdEncoding.EncodeToString(ciphertext)
	return nil
}

func (v *LocalEncryptedVault) GetPrivateKey(hostAlias string) (string, error) {
	encStr, exists := v.KeyStore[hostAlias]
	if !exists {
		return "", fmt.Errorf("context: %w", errors.New("private key missing for target alias"))
	}
	
	data, err := base64.StdEncoding.DecodeString(encStr)
	if err != nil {
		return "", fmt.Errorf("context: %w", err)
	}
	
	block, err := aes.NewCipher(v.MasterKey)
	if err != nil {
		return "", fmt.Errorf("context: %w", err)
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("context: %w", err)
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("context: %w", errors.New("ciphertext payload malformed"))
	}
	
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("context: %w", err)
	}
	
	return string(plaintext), nil
}

func (v *LocalEncryptedVault) EncryptAndStoreSudoPassword(hostAlias, password string) error {
	return v.EncryptAndStore(hostAlias+sudoKeySuffix, password)
}

func (v *LocalEncryptedVault) GetSudoPassword(hostAlias string) (string, error) {
	plaintext, err := v.GetPrivateKey(hostAlias + sudoKeySuffix)
	if err != nil {
		return "", fmt.Errorf("context: %w", ErrSudoPasswordMissing)
	}
	return plaintext, nil
}

// StoreAPIKey encrypts the LLM API key at rest in the vault.
func (v *LocalEncryptedVault) StoreAPIKey(apiKey string) error {
	return v.EncryptAndStore(apiKeyAlias, apiKey)
}

// GetAPIKey returns the decrypted LLM API key, or ErrAPIKeyMissing if none is
// configured.
func (v *LocalEncryptedVault) GetAPIKey() (string, error) {
	plaintext, err := v.GetPrivateKey(apiKeyAlias)
	if err != nil {
		return "", fmt.Errorf("context: %w", ErrAPIKeyMissing)
	}
	return plaintext, nil
}

// ApplyAPIKeyEnv seeds the LLM API key from the process environment
// (OPENAI_API_KEY) into the encrypted vault, so the key is encrypted at rest
// and resolved from the vault rather than read directly from the environment.
func ApplyAPIKeyEnv(v *LocalEncryptedVault, environ []string) {
	for _, kv := range environ {
		idx := strings.IndexByte(kv, '=')
		if idx < 0 {
			continue
		}
		key, val := kv[:idx], kv[idx+1:]
		if key != "OPENAI_API_KEY" {
			continue
		}
		if val == "" {
			return
		}
		_ = v.StoreAPIKey(val)
		return
	}
}

// Save writes the (already-encrypted) KeyStore to disk so secrets survive
// process restarts. The master key is never written; only ciphertext is
// persisted.
func (v *LocalEncryptedVault) Save(path string) error {
	data, err := json.Marshal(v.KeyStore)
	if err != nil {
		return fmt.Errorf("context: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("context: %w", err)
	}
	return nil
}

// Load reads a previously saved encrypted KeyStore from disk. A missing file
// is treated as a first-run bootstrap (no secrets yet) and returns no error.
func (v *LocalEncryptedVault) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("context: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	var store map[string]string
	if err := json.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("context: %w", err)
	}
	if v.KeyStore == nil {
		v.KeyStore = make(map[string]string)
	}
	for k, val := range store {
		v.KeyStore[k] = val
	}
	return nil
}

// VaultPath returns the conventional on-disk vault location relative to the
// current working directory.
func VaultPath() string {
	return vaultFileName
}

// sudoPassEnvPrefix mirrors Ansible's per-host ansible_become_password, supplied
// via environment (e.g. SUDO_PASS_WEB_PROD_01) and stored encrypted in the vault.
const sudoPassEnvPrefix = "SUDO_PASS_"

// ApplySudoPasswordEnv seeds per-host sudo passwords from the process
// environment into the encrypted vault, so secrets are encrypted at rest and
// resolved per host (never a single shared password). The env key is the prefix
// plus the upper-cased host alias; the alias is recovered by lower-casing.
func ApplySudoPasswordEnv(v *LocalEncryptedVault, environ []string) {
	for _, kv := range environ {
		idx := strings.IndexByte(kv, '=')
		if idx < 0 {
			continue
		}
		key, val := kv[:idx], kv[idx+1:]
		if !strings.HasPrefix(key, sudoPassEnvPrefix) {
			continue
		}
		alias := strings.ToLower(strings.TrimPrefix(key, sudoPassEnvPrefix))
		// Env vars can't contain dashes, so callers use underscores; map them
		// back to dashes to match inventory host aliases (web_prod_01 -> web-prod-01).
		alias = strings.ReplaceAll(alias, "_", "-")
		if alias == "" || val == "" {
			continue
		}
		_ = v.EncryptAndStoreSudoPassword(alias, val)
	}
}

var _ agent.SecretVault = (*LocalEncryptedVault)(nil)
