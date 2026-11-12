package controller

import "time"

// SessionInputDTO is the body of POST /api/focus/sessions: a finished
// focus phase.
type SessionInputDTO struct {
	Minutes int       `json:"minutes"`
	EndedAt time.Time `json:"endedAt"`
}
