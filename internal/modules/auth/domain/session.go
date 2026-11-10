package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"trisse-home/internal/shared/kernel"
)

// sessionRefreshInterval limits how often a sliding session is extended so
// that not every request causes a database write.
const sessionRefreshInterval = time.Hour

// TokenHash is the SHA-256 hash of a session token. Only the hash is stored,
// so a leaked database does not reveal valid session cookies.
type TokenHash string

// HashToken returns the TokenHash of a raw session token.
func HashToken(raw string) TokenHash {
	sum := sha256.Sum256([]byte(raw))
	return TokenHash(hex.EncodeToString(sum[:]))
}

// Session is an authenticated browser session with a sliding expiry.
type Session struct {
	tokenHash   TokenHash
	csrfToken   string
	createdAt   time.Time
	expiresAt   time.Time
	refreshedAt time.Time
}

// NewSession starts a session that expires after ttl.
func NewSession(tokenHash TokenHash, csrfToken string, now time.Time, ttl time.Duration) (Session, error) {
	if tokenHash == "" || csrfToken == "" {
		return Session{}, kernel.NewValidationError("session", "token and csrf token are required")
	}
	if ttl <= 0 {
		return Session{}, kernel.NewValidationError("ttl", "must be positive")
	}
	return Session{tokenHash: tokenHash, csrfToken: csrfToken, createdAt: now, expiresAt: now.Add(ttl), refreshedAt: now}, nil
}

// RestoreSession rebuilds a session from persisted state.
func RestoreSession(tokenHash TokenHash, csrfToken string, createdAt, expiresAt, refreshedAt time.Time) Session {
	return Session{tokenHash: tokenHash, csrfToken: csrfToken, createdAt: createdAt, expiresAt: expiresAt, refreshedAt: refreshedAt}
}

// TokenHash returns the session identifier.
func (s Session) TokenHash() TokenHash { return s.tokenHash }

// CSRFToken returns the per-session anti-CSRF token.
func (s Session) CSRFToken() string { return s.csrfToken }

// CreatedAt returns the login time.
func (s Session) CreatedAt() time.Time { return s.createdAt }

// ExpiresAt returns the expiry time.
func (s Session) ExpiresAt() time.Time { return s.expiresAt }

// RefreshedAt returns when the expiry was last extended.
func (s Session) RefreshedAt() time.Time { return s.refreshedAt }

// IsExpired reports whether the session is no longer valid at now.
func (s Session) IsExpired(now time.Time) bool { return !now.Before(s.expiresAt) }

// Refresh extends the expiry to now+ttl if the last extension is older than
// the refresh interval. It reports whether the session changed.
func (s *Session) Refresh(now time.Time, ttl time.Duration) bool {
	if now.Sub(s.refreshedAt) < sessionRefreshInterval {
		return false
	}
	s.expiresAt, s.refreshedAt = now.Add(ttl), now
	return true
}
