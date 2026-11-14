// Package presenter maps holidays to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/holidays/domain"
	"trisse-home/internal/modules/holidays/usecase"
)

// EntryDTO is one public holiday or school holiday period; dates are YYYY-MM-DD.
type EntryDTO struct {
	// Kind is "public" or "school".
	Kind string `json:"kind"`
	// Key names a public holiday (translated by the client), Name a school holiday period.
	Key       string `json:"key"`
	Name      string `json:"name"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Current   bool   `json:"current"`
	DaysUntil int    `json:"daysUntil"`
}

// OverviewDTO is returned by GET /api/holidays.
type OverviewDTO struct {
	State           string     `json:"state"`
	Entries         []EntryDTO `json:"entries"`
	School          bool       `json:"school"`
	SchoolUpdatedAt *time.Time `json:"schoolUpdatedAt"`
	SchoolError     string     `json:"schoolError"`
}

// SettingsDTO is returned by the settings endpoints.
type SettingsDTO struct {
	SchoolURL string `json:"schoolUrl"`
}

// Overview maps an Overview.
func Overview(o usecase.Overview) OverviewDTO {
	dto := OverviewDTO{State: string(o.State), School: o.School, SchoolError: o.SchoolError, Entries: make([]EntryDTO, 0, len(o.Entries))}
	if !o.SchoolUpdatedAt.IsZero() {
		at := o.SchoolUpdatedAt
		dto.SchoolUpdatedAt = &at
	}
	for _, e := range o.Entries {
		dto.Entries = append(dto.Entries, EntryDTO{Kind: string(e.Kind), Key: e.Key, Name: e.Name,
			Start: e.Start.Format(time.DateOnly), End: e.End.Format(time.DateOnly), Current: e.Current, DaysUntil: e.DaysUntil})
	}
	return dto
}

// Settings maps the settings.
func Settings(s domain.Settings) SettingsDTO { return SettingsDTO{SchoolURL: s.SchoolURL} }
