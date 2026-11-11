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
	publicID    string
	csrfToken   string
	client      Client
	createdAt   time.Time
	expiresAt   time.Time
	refreshedAt time.Time
}

// maxUserAgentLength bounds the stored user agent.
const maxUserAgentLength = 300

// Client describes the browser a session was created from. It is shown in
// the session overview so the user can recognise their devices.
type Client struct {
	UserAgent string
	IP        string
}

// NewClient normalises client metadata.
func NewClient(userAgent, ip string) Client {
	if len(userAgent) > maxUserAgentLength {
		userAgent = userAgent[:maxUserAgentLength]
	}
	return Client{UserAgent: userAgent, IP: ip}
}

// NewSession starts a session that expires after ttl. publicID identifies
// the session in the session overview; unlike the token hash it may be shown.
func NewSession(tokenHash TokenHash, publicID, csrfToken string, client Client, now time.Time, ttl time.Duration) (Session, error) {
	if tokenHash == "" || csrfToken == "" || publicID == "" {
		return Session{}, kernel.NewValidationError("session", "token, public id and csrf token are required")
	}
	if ttl <= 0 {
		return Session{}, kernel.NewValidationError("ttl", "must be positive")
	}
	return Session{tokenHash: tokenHash, publicID: publicID, csrfToken: csrfToken, client: client,
		createdAt: now, expiresAt: now.Add(ttl), refreshedAt: now}, nil
}

// RestoreSession rebuilds a session from persisted state.
func RestoreSession(tokenHash TokenHash, publicID, csrfToken string, client Client, createdAt, expiresAt, refreshedAt time.Time) Session {
	return Session{tokenHash: tokenHash, publicID: publicID, csrfToken: csrfToken, client: client,
		createdAt: createdAt, expiresAt: expiresAt, refreshedAt: refreshedAt}
}

// PublicID returns the identifier shown in the session overview.
func (s Session) PublicID() string { return s.publicID }

// Client returns the browser metadata.
func (s Session) Client() Client { return s.client }

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
