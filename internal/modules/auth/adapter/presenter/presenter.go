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
