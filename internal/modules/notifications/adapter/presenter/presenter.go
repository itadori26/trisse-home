// Package presenter maps notification settings to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/notifications/usecase"
)

// RulesDTO selects what is notified.
type RulesDTO struct {
	Services      bool    `json:"services"`
	Containers    bool    `json:"containers"`
	Disk          bool    `json:"disk"`
	DiskThreshold float64 `json:"diskThreshold"`
	TLS           bool    `json:"tls"`
	TLSDays       int     `json:"tlsDays"`
	Backups       bool    `json:"backups"`
	Security      bool    `json:"security"`
}

// SettingsDTO is returned by the settings endpoints; the token is never returned.
type SettingsDTO struct {
	Enabled       bool       `json:"enabled"`
	ServerURL     string     `json:"serverUrl"`
	Topic         string     `json:"topic"`
	HasToken      bool       `json:"hasToken"`
	Rules         RulesDTO   `json:"rules"`
	LastError     string     `json:"lastError"`
	LastDelivered *time.Time `json:"lastDelivered"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO {
	dto := SettingsDTO{Enabled: v.Enabled, ServerURL: v.ServerURL, Topic: v.Topic, HasToken: v.HasToken,
		Rules: RulesDTO(v.Rules), LastError: v.LastError}
	if !v.LastDelivered.IsZero() {
		t := v.LastDelivered
		dto.LastDelivered = &t
	}
	return dto
}
