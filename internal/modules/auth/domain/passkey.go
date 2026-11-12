package domain

import (
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// MaxPasskeys bounds the passkeys per account.
const MaxPasskeys = 10

// Passkey errors.
var (
	ErrPasskeyNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "The passkey does not exist.")
	ErrPasskeyCloned      = kernel.NewPublicError(kernel.ErrUnauthorized, "This passkey reported an old signature counter and may have been cloned. Remove it and register it again.")
	ErrTooManyPasskeys    = kernel.NewPublicError(kernel.ErrInvalid, "There are already 10 passkeys.")
	ErrPasskeysNeedHTTPS  = kernel.NewPublicError(kernel.ErrNotFound, "Passkeys need HTTPS and PUBLIC_URL (see the README).")
	ErrPasskeyChallenge   = kernel.NewPublicError(kernel.ErrUnauthorized, "The passkey request expired. Please try again.")
	ErrPasskeyUnsupported = kernel.NewPublicError(kernel.ErrInvalid, "This passkey could not be verified.")
)

// Passkey is a WebAuthn credential registered as a second factor.
type Passkey struct {
	ID           string
	AccountID    string
	CredentialID []byte
	PublicKey    []byte // PKIX DER
	Alg          int
	SignCount    uint32
	Name         string
	CreatedAt    time.Time
	LastUsedAt   time.Time
}

// NewPasskey validates a newly registered passkey.
func NewPasskey(id, accountID string, credentialID, publicKey []byte, alg int, signCount uint32, name string, now time.Time) (Passkey, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Passkey"
	}
	if len(name) > 60 {
		return Passkey{}, kernel.NewValidationError("name", "The name must be at most 60 characters long.")
	}
	if id == "" || accountID == "" || len(credentialID) == 0 || len(publicKey) == 0 {
		return Passkey{}, kernel.NewValidationError("passkey", "incomplete credential")
	}
	return Passkey{ID: id, AccountID: accountID, CredentialID: credentialID, PublicKey: publicKey, Alg: alg,
		SignCount: signCount, Name: name, CreatedAt: now}, nil
}

// Use records a successful assertion. A counter that does not increase
// although the authenticator uses counters indicates a cloned
// authenticator (WebAuthn §6.1.1); authenticators without counters always
// report 0 and are accepted.
func (p *Passkey) Use(signCount uint32, now time.Time) error {
	if (signCount != 0 || p.SignCount != 0) && signCount <= p.SignCount {
		return ErrPasskeyCloned
	}
	p.SignCount, p.LastUsedAt = signCount, now
	return nil
}
