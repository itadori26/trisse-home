// Package presenter maps the briefing to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/briefing/domain"
	"trisse-home/internal/modules/briefing/usecase"
)

// ItemDTO is one line of a section. severity is "info", "warning" or
// "critical"; at is RFC 3339 or empty.
type ItemDTO struct {
	Text     string `json:"text"`
	Detail   string `json:"detail"`
	At       string `json:"at"`
	Severity string `json:"severity"`
}

// SectionDTO is a non-empty section; more counts the items not shown.
type SectionDTO struct {
	Kind  string    `json:"kind"`
	Items []ItemDTO `json:"items"`
	More  int       `json:"more"`
}

// BriefingDTO is returned by GET /api/briefing.
type BriefingDTO struct {
	Sections []SectionDTO `json:"sections"`
	// Unavailable lists sections whose source could not be read.
	Unavailable []string  `json:"unavailable"`
	At          time.Time `json:"at"`
}

// LayoutEntryDTO switches a section on or off.
type LayoutEntryDTO struct {
	Kind    string `json:"kind"`
	Enabled bool   `json:"enabled"`
}

// SettingsDTO is returned by GET/PUT /api/briefing/settings.
type SettingsDTO struct {
	Layout    []LayoutEntryDTO `json:"layout"`
	PushOn    bool             `json:"pushOn"`
	PushAt    string           `json:"pushAt"`
	PushDays  []int            `json:"pushDays"`
	SkipEmpty bool             `json:"skipEmpty"`
	Lat       float64          `json:"lat"`
	Lon       float64          `json:"lon"`
	// Available lists the sections whose source module is enabled.
	Available []string `json:"available"`
}

var severities = map[domain.Severity]string{
	domain.SeverityInfo: "info", domain.SeverityWarning: "warning", domain.SeverityCritical: "critical",
}

// Briefing maps a BriefingView.
func Briefing(v usecase.BriefingView) BriefingDTO {
	out := BriefingDTO{Sections: []SectionDTO{}, Unavailable: []string{}, At: v.At}
	for _, s := range v.Briefing.Sections {
		dto := SectionDTO{Kind: string(s.Kind), More: s.More, Items: make([]ItemDTO, 0, len(s.Items))}
		for _, it := range s.Items {
			item := ItemDTO{Text: it.Text, Detail: it.Detail, Severity: severities[it.Severity]}
			if !it.At.IsZero() {
				item.At = it.At.Format(time.RFC3339)
			}
			dto.Items = append(dto.Items, item)
		}
		out.Sections = append(out.Sections, dto)
	}
	for _, k := range v.Unavailable {
		out.Unavailable = append(out.Unavailable, string(k))
	}
	return out
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO {
	out := SettingsDTO{PushOn: v.PushOn, PushAt: v.PushAt, PushDays: append([]int{}, v.PushDays...), SkipEmpty: v.SkipEmpty,
		Lat: v.Lat, Lon: v.Lon, Layout: make([]LayoutEntryDTO, 0, len(v.Layout)), Available: make([]string, 0, len(v.Available))}
	for _, e := range v.Layout {
		out.Layout = append(out.Layout, LayoutEntryDTO{Kind: string(e.Kind), Enabled: e.Enabled})
	}
	for _, k := range v.Available {
		out.Available = append(out.Available, string(k))
	}
	return out
}
