// Package presenter maps the FRITZ!Box status to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/fritzbox/usecase"
)

// SettingsDTO is returned by the settings endpoints (never the password).
type SettingsDTO struct {
	Configured  bool   `json:"configured"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword"`
}

// WANDTO is the internet connection; rates in bytes/s, link speeds in bits/s.
type WANDTO struct {
	Connected         bool    `json:"connected"`
	ExternalIP        string  `json:"externalIp"`
	UptimeSeconds     int64   `json:"uptimeSeconds"`
	ThroughputKnown   bool    `json:"throughputKnown"`
	DownstreamBs      int64   `json:"downstreamBytesPerSecond"`
	UpstreamBs        int64   `json:"upstreamBytesPerSecond"`
	DownstreamMaxBits int64   `json:"downstreamMaxBits"`
	UpstreamMaxBits   int64   `json:"upstreamMaxBits"`
	DownstreamPercent float64 `json:"downstreamPercent"`
	UpstreamPercent   float64 `json:"upstreamPercent"`
}

// CallDTO is one call.
type CallDTO struct {
	Type            string    `json:"type" tstype:"'incoming' | 'missed' | 'outgoing' | 'rejected' | 'active'"`
	Number          string    `json:"number"`
	Name            string    `json:"name"`
	At              time.Time `json:"at"`
	DurationMinutes int       `json:"durationMinutes"`
}

// StatusDTO is returned by GET /api/fritzbox/status.
type StatusDTO struct {
	Model     string    `json:"model"`
	WAN       WANDTO    `json:"wan"`
	Calls     []CallDTO `json:"calls"`
	Hint      string    `json:"hint"`
	UpdatedAt time.Time `json:"updatedAt"`
	Error     string    `json:"error"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Status maps a StatusView.
func Status(v usecase.StatusView) StatusDTO {
	s, w := v.Status, v.Status.WAN
	dto := StatusDTO{
		Model: s.Model, Hint: s.Hint, UpdatedAt: v.UpdatedAt, Error: v.Error, Calls: make([]CallDTO, 0, len(s.Calls)),
		WAN: WANDTO{
			Connected: w.Connected, ExternalIP: w.ExternalIP, UptimeSeconds: int64(w.Uptime.Seconds()), ThroughputKnown: w.Known,
			DownstreamBs: w.DownstreamBs, UpstreamBs: w.UpstreamBs, DownstreamMaxBits: w.DownstreamMaxBits, UpstreamMaxBits: w.UpstreamMaxBits,
			DownstreamPercent: w.DownstreamPercent(), UpstreamPercent: w.UpstreamPercent(),
		},
	}
	for _, c := range s.Calls {
		dto.Calls = append(dto.Calls, CallDTO{Type: string(c.Type), Number: c.Number, Name: c.Name, At: c.At, DurationMinutes: int(c.Duration.Minutes())})
	}
	return dto
}
