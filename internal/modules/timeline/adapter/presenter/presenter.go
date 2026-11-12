// Package presenter maps timeline events to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/timeline/domain"
)

// EventDTO is one timeline entry.
type EventDTO struct {
	ID       string    `json:"id"`
	At       time.Time `json:"at"`
	Source   string    `json:"source"`
	Severity string    `json:"severity" tstype:"'info' | 'warning' | 'critical'"`
	Subject  string    `json:"subject"`
	Message  string    `json:"message"`
}

// Events maps events.
func Events(list []domain.Event) []EventDTO {
	out := make([]EventDTO, 0, len(list))
	for _, e := range list {
		out = append(out, EventDTO{ID: e.ID, At: e.At, Source: e.Source, Severity: string(e.Severity), Subject: e.Subject, Message: e.Message})
	}
	return out
}

// SettingsDTO is the timeline configuration.
type SettingsDTO struct {
	RetentionDays int `json:"retentionDays"`
}
