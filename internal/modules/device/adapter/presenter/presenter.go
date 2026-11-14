// Package presenter maps the display state to JSON.
//
// Ring: Interface Adapters.
package presenter

import "trisse-home/internal/modules/device/domain"

// StateDTO is returned by GET /api/device (#140).
type StateDTO struct {
	// Requested is "", "headless" or "kiosk" ("" = not chosen yet).
	Requested string `json:"requested" tstype:"'' | 'headless' | 'kiosk'"`
	Phase     string `json:"phase" tstype:"'idle' | 'pending' | 'installing' | 'active' | 'failed'"`
	Error     string `json:"error"`
	Display   string `json:"display" tstype:"'connected' | 'disconnected' | 'unknown'"`
}

// State maps the state.
func State(s domain.State) StateDTO {
	return StateDTO{Requested: string(s.Requested), Phase: string(s.Phase()), Error: s.Progress.Error, Display: string(s.Display)}
}
