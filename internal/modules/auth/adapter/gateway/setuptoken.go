package gateway

import "crypto/subtle"

// SetupToken verifies the one-time setup token in constant time.
type SetupToken struct{ token string }

// NewSetupToken returns a verifier for token.
func NewSetupToken(token string) *SetupToken { return &SetupToken{token: token} }

// Verify implements usecase.SetupTokenVerifier.
func (s *SetupToken) Verify(candidate string) bool {
	if s.token == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(s.token), []byte(candidate)) == 1
}
