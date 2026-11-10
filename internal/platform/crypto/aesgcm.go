package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// ErrDecrypt reports a ciphertext that is corrupt or was sealed with another key.
var ErrDecrypt = errors.New("decrypt secret")

// AESGCM encrypts small secrets (API keys, passwords of integrations) before
// they are stored in the database.
type AESGCM struct {
	aead cipher.AEAD
}

// NewAESGCM returns an AESGCM using a 32-byte key.
func NewAESGCM(key []byte) (*AESGCM, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &AESGCM{aead: aead}, nil
}

// Encrypt returns base64(nonce || ciphertext).
func (a *AESGCM) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := a.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt reverses Encrypt.
func (a *AESGCM) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(raw) < a.aead.NonceSize() {
		return "", ErrDecrypt
	}
	nonce, ciphertext := raw[:a.aead.NonceSize()], raw[a.aead.NonceSize():]
	plain, err := a.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecrypt
	}
	return string(plain), nil
}
