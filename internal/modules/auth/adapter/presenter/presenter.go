// Package presenter converts auth use case responses into the JSON DTOs of
// the HTTP API. The DTO types are the source for the generated TypeScript
// types of the frontend.
//
// Ring: Interface Adapters.
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
