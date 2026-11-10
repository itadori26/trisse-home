package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// TokenBytes is the entropy of generated tokens (256 bit).
const TokenBytes = 32

// RandomTokens generates URL-safe random tokens.
type RandomTokens struct{}

// NewToken returns a new random URL-safe token.
func (RandomTokens) NewToken() (string, error) {
	b := make([]byte, TokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
