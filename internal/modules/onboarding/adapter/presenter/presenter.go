// Package presenter maps the onboarding state to its JSON DTO.
//
// Ring: Interface Adapters.
package presenter

import "trisse-home/internal/modules/onboarding/usecase"

// StatusDTO is returned by the onboarding endpoints.
type StatusDTO struct {
	Completed bool `json:"completed"`
}

// Status maps a StatusResponse.
func Status(r usecase.StatusResponse) StatusDTO { return StatusDTO{Completed: r.Completed} }
