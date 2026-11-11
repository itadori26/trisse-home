package presenter

import (
	"time"

	"trisse-home/internal/modules/auth/usecase"
)

// StatusDTO is returned by GET /api/auth/status.
type StatusDTO struct {
	SetupRequired bool   `json:"setupRequired"`
	Authenticated bool   `json:"authenticated"`
	CSRFToken     string `json:"csrfToken,omitempty"`
	// Demo is true on a public demo instance (POST /api/auth/demo logs in).
	Demo bool `json:"demo"`
}

// SessionDTO is returned after setup, login and password change. The session
// token itself is only sent as an HttpOnly cookie.
type SessionDTO struct {
	CSRFToken string    `json:"csrfToken"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Status maps a StatusResponse to its DTO.
func Status(r usecase.StatusResponse) StatusDTO {
	return StatusDTO{SetupRequired: r.SetupRequired, Authenticated: r.Authenticated, CSRFToken: r.CSRFToken}
}

// Session maps a SessionResponse to its DTO.
func Session(r usecase.SessionResponse) SessionDTO {
	return SessionDTO{CSRFToken: r.CSRFToken, ExpiresAt: r.ExpiresAt}
}

// SessionInfoDTO describes an active session in the session overview.
type SessionInfoDTO struct {
	ID         string    `json:"id"`
	UserAgent  string    `json:"userAgent"`
	ClientIP   string    `json:"clientIp"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Current    bool      `json:"current"`
}

// RevokedDTO is returned by POST /api/auth/sessions/revoke-others.
type RevokedDTO struct {
	Revoked int64 `json:"revoked"`
}

// Sessions maps session views.
func Sessions(vs []usecase.SessionView) []SessionInfoDTO {
	out := make([]SessionInfoDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, SessionInfoDTO(v))
	}
	return out
}

// TwoFactorStatusDTO is returned by GET /api/auth/totp.
type TwoFactorStatusDTO struct {
	Enabled           bool `json:"enabled"`
	RecoveryCodesLeft int  `json:"recoveryCodesLeft"`
}

// EnrollmentDTO is returned by POST /api/auth/totp/setup; uri is encoded as QR code.
type EnrollmentDTO struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

// RecoveryCodesDTO is returned once by POST /api/auth/totp/confirm.
type RecoveryCodesDTO struct {
	Codes []string `json:"codes"`
}

// SecurityEventDTO is one entry of GET /api/auth/events.
type SecurityEventDTO struct {
	ID        string    `json:"id"`
	At        time.Time `json:"at"`
	Type      string    `json:"type"`
	AccountID string    `json:"accountId"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"userAgent"`
	Detail    string    `json:"detail"`
}

// SecurityEvents maps the security log.
func SecurityEvents(list []usecase.SecurityEventView) []SecurityEventDTO {
	out := make([]SecurityEventDTO, 0, len(list))
	for _, e := range list {
		out = append(out, SecurityEventDTO(e))
	}
	return out
}
