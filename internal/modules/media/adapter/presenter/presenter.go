// Package presenter maps media use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"
	"time"

	"trisse-home/internal/modules/media/domain"
	"trisse-home/internal/modules/media/usecase"
)

// SettingsDTO is returned by GET/PUT /api/media/settings; the token is
// never returned.
type SettingsDTO struct {
	Configured bool   `json:"configured"`
	Provider   string `json:"provider" tstype:"'jellyfin' | 'plex'"`
	URL        string `json:"url"`
	HasToken   bool   `json:"hasToken"`
}

// ItemDTO is a recently added item; the poster is at /api/media/posters/{id}.
type ItemDTO struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Series  string    `json:"series"`
	Kind    string    `json:"kind"`
	Year    int       `json:"year"`
	AddedAt time.Time `json:"addedAt"`
}

// SessionDTO is something that is being played.
type SessionDTO struct {
	User     string  `json:"user"`
	Device   string  `json:"device"`
	Title    string  `json:"title"`
	Series   string  `json:"series"`
	Kind     string  `json:"kind"`
	Paused   bool    `json:"paused"`
	Progress float64 `json:"progress"`
}

// OverviewDTO is returned by GET /api/media/overview.
type OverviewDTO struct {
	ServerName string       `json:"serverName"`
	Recent     []ItemDTO    `json:"recent"`
	Sessions   []SessionDTO `json:"sessions"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Overview maps the overview.
func Overview(o domain.Overview) OverviewDTO {
	dto := OverviewDTO{ServerName: o.ServerName, Recent: make([]ItemDTO, 0, len(o.Recent)), Sessions: make([]SessionDTO, 0, len(o.Sessions))}
	for _, it := range o.Recent {
		dto.Recent = append(dto.Recent, ItemDTO{ID: it.ID, Title: it.Title, Series: it.Series, Kind: it.Kind, Year: it.Year, AddedAt: it.AddedAt})
	}
	for _, s := range o.Sessions {
		dto.Sessions = append(dto.Sessions, SessionDTO{User: s.User, Device: s.Device, Title: s.Title, Series: s.Series, Kind: s.Kind,
			Paused: s.Paused, Progress: math.Round(s.Progress())})
	}
	return dto
}
