package domain

import (
	"time"

	"trisse-home/internal/shared/kernel"
)

// EventType classifies a security event. The values are stable: they are
// stored and translated in the frontend.
type EventType string

// Security event types.
const (
	EventLoginSucceeded    EventType = "login_succeeded"
	EventLoginFailed       EventType = "login_failed"
	EventLoginLocked       EventType = "login_locked"
	EventNewDevice         EventType = "new_device"
	EventLogoutEverywhere  EventType = "logout_everywhere"
	EventPasswordChanged   EventType = "password_changed"
	EventTwoFactorEnabled  EventType = "two_factor_enabled"
	EventTwoFactorDisabled EventType = "two_factor_disabled"
	EventRecoveryCodeUsed  EventType = "recovery_code_used"
	EventSessionRevoked    EventType = "session_revoked"
	EventGuestChanged      EventType = "guest_changed"
	EventTokenCreated      EventType = "token_created"
	EventTokenRevoked      EventType = "token_revoked"
	EventTokenUsed         EventType = "token_used"
	EventPasskeyAdded      EventType = "passkey_added"
	EventPasskeyRemoved    EventType = "passkey_removed"
	EventExternal          EventType = "external"
)

// EventRetention is how long security events are kept. IP addresses are
// personal data, so the log is short-lived.
const EventRetention = 90 * 24 * time.Hour

// maxEventDetail bounds free text stored with an event.
const maxEventDetail = 200

// SecurityEvent is one entry of the security log. It never contains
// passwords, codes or tokens.
type SecurityEvent struct {
	ID        string
	At        time.Time
	Type      EventType
	AccountID string
	Client    Client
	Detail    string
}

// NewSecurityEvent creates an event; detail is truncated.
func NewSecurityEvent(id string, at time.Time, t EventType, accountID string, client Client, detail string) (SecurityEvent, error) {
	if id == "" || t == "" {
		return SecurityEvent{}, kernel.NewValidationError("event", "id and type are required")
	}
	if len(detail) > maxEventDetail {
		detail = detail[:maxEventDetail]
	}
	return SecurityEvent{ID: id, At: at, Type: t, AccountID: accountID, Client: client, Detail: detail}, nil
}

// LoginThrottle limits failed attempts per client: after MaxFailures within
// Window the client is locked out, and every further lockout doubles the
// lockout time up to MaxLockout (exponential backoff).
type LoginThrottle struct {
	MaxFailures int
	Window      time.Duration
	BaseLockout time.Duration
	MaxLockout  time.Duration
}

// DefaultLoginThrottle allows 5 failures per 15 minutes, then locks for
// 15 minutes, 30, 60 … up to 24 hours.
var DefaultLoginThrottle = LoginThrottle{MaxFailures: 5, Window: 15 * time.Minute, BaseLockout: 15 * time.Minute, MaxLockout: 24 * time.Hour}

// Lockout returns the lockout duration for the n-th lockout (0-based).
func (t LoginThrottle) Lockout(n int) time.Duration {
	d := t.BaseLockout
	for i := 0; i < n && d < t.MaxLockout; i++ {
		d *= 2
	}
	return min(d, t.MaxLockout)
}
