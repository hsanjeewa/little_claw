package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/devops/agent/internal/domain/agent"
)

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

func (v *LocalEncryptedVault) GetSudoPassword(hostAlias string) (string, error) {
	return "mocked-secure-sudo-password", nil
}

var _ agent.SecretVault = (*LocalEncryptedVault)(nil)
