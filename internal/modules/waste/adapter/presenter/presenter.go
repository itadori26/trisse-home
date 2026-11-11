// Package presenter maps waste collections to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/modules/waste/usecase"
)

// RuleDTO maps summaries to a label and style.
type RuleDTO struct {
	Match string `json:"match"`
	Label string `json:"label"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

// SettingsDTO is returned by the settings endpoints.
type SettingsDTO struct {
	Configured bool      `json:"configured"`
	URL        string    `json:"url"`
	Rules      []RuleDTO `json:"rules"`
}

// TypeDTO is one waste type with its next dates (YYYY-MM-DD).
type TypeDTO struct {
	Label     string   `json:"label"`
	Color     string   `json:"color"`
	Icon      string   `json:"icon"`
	Dates     []string `json:"dates"`
	DaysUntil int      `json:"daysUntil"`
}

// OverviewDTO is returned by GET /api/waste/upcoming.
type OverviewDTO struct {
	Types     []TypeDTO `json:"types"`
	UpdatedAt time.Time `json:"updatedAt"`
	Error     string    `json:"error"`
}

// Settings maps the settings.
func Settings(s domain.Settings, configured bool) SettingsDTO {
	dto := SettingsDTO{Configured: configured, URL: s.URL, Rules: make([]RuleDTO, 0, len(s.Rules))}
	for _, r := range s.Rules {
		dto.Rules = append(dto.Rules, RuleDTO(r))
	}
	return dto
}

// Overview maps an Overview.
func Overview(o usecase.Overview) OverviewDTO {
	dto := OverviewDTO{UpdatedAt: o.UpdatedAt, Error: o.Error, Types: make([]TypeDTO, 0, len(o.Types))}
	for _, u := range o.Types {
		t := TypeDTO{Label: u.Style.Label, Color: u.Style.Color, Icon: u.Style.Icon, DaysUntil: u.DaysUntil}
		for _, d := range u.Dates {
			t.Dates = append(t.Dates, d.Format(time.DateOnly))
		}
		dto.Types = append(dto.Types, t)
	}
	return dto
}
