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

// CalendarFileDTO describes an uploaded ICS file (#153); dates are YYYY-MM-DD.
type CalendarFileDTO struct {
	Name       string    `json:"name"`
	UploadedAt time.Time `json:"uploadedAt"`
	Events     int       `json:"events"`
	From       string    `json:"from"`
	To         string    `json:"to"`
}

// SettingsDTO is returned by the settings endpoints.
type SettingsDTO struct {
	Configured bool `json:"configured"`
	// Source is "url", "file" or "" (none).
	Source string `json:"source"`
	URL    string `json:"url"`
	// File is set for the file source; the content is never returned.
	File  *CalendarFileDTO `json:"file"`
	Rules []RuleDTO        `json:"rules"`
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
	// EndsOn (YYYY-MM-DD) is set when an uploaded calendar ends within 30 days.
	EndsOn string `json:"endsOn"`
}

// Settings maps the settings.
func Settings(s domain.Settings, configured bool) SettingsDTO {
	dto := SettingsDTO{Configured: configured, Source: string(s.Source.Kind), URL: s.Source.URL, Rules: make([]RuleDTO, 0, len(s.Rules))}
	if s.Source.Kind == domain.SourceFile {
		f := s.Source.File
		dto.File = &CalendarFileDTO{Name: f.Name, UploadedAt: f.UploadedAt, Events: f.Events,
			From: f.From.Format(time.DateOnly), To: f.To.Format(time.DateOnly)}
	}
	for _, r := range s.Rules {
		dto.Rules = append(dto.Rules, RuleDTO(r))
	}
	return dto
}

// Overview maps an Overview.
func Overview(o usecase.Overview) OverviewDTO {
	dto := OverviewDTO{UpdatedAt: o.UpdatedAt, Error: o.Error, Types: make([]TypeDTO, 0, len(o.Types))}
	if !o.EndsOn.IsZero() {
		dto.EndsOn = o.EndsOn.Format(time.DateOnly)
	}
	for _, u := range o.Types {
		t := TypeDTO{Label: u.Style.Label, Color: u.Style.Color, Icon: u.Style.Icon, DaysUntil: u.DaysUntil}
		for _, d := range u.Dates {
			t.Dates = append(t.Dates, d.Format(time.DateOnly))
		}
		dto.Types = append(dto.Types, t)
	}
	return dto
}
