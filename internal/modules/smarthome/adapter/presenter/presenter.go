// Package presenter maps smart home use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/smarthome/usecase"
)

// SettingsDTO is returned by GET/PUT /api/smarthome/settings; the token is
// never returned.
type SettingsDTO struct {
	Configured bool   `json:"configured"`
	URL        string `json:"url"`
	HasToken   bool   `json:"hasToken"`
}

// EntityDTO is the state of one entity.
type EntityDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	State       string    `json:"state"`
	Unit        string    `json:"unit"`
	LastChanged time.Time `json:"lastChanged"`
	// Action is empty for read-only entities and those not allow-listed.
	Action string `json:"action" tstype:"'' | 'toggle' | 'press'"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Entities maps entity views.
func Entities(list []usecase.EntityView) []EntityDTO {
	out := make([]EntityDTO, 0, len(list))
	for _, e := range list {
		out = append(out, EntityDTO{ID: e.ID, Name: e.Name, State: e.State, Unit: e.Unit, LastChanged: e.LastChanged, Action: string(e.Action)})
	}
	return out
}
