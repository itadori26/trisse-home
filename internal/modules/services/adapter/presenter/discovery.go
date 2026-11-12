package presenter

import (
	"time"

	"trisse-home/internal/modules/services/usecase"
)

// DiscoveryDTO is the state of Docker autodiscovery (#78).
type DiscoveryDTO struct {
	// Available is false when the docker module is not configured.
	Available       bool   `json:"available"`
	Enabled         bool   `json:"enabled"`
	IntervalMinutes int    `json:"intervalMinutes"`
	Rule            string `json:"rule" tstype:"'labels' | 'manual'"`
	// LastSync is null before the first sync.
	LastSync *SyncResultDTO `json:"lastSync"`
}

// SyncResultDTO is the outcome of the last sync.
type SyncResultDTO struct {
	At      time.Time `json:"at"`
	Error   string    `json:"error"`
	Created int       `json:"created"`
	Updated int       `json:"updated"`
	Hidden  int       `json:"hidden"`
	Shown   int       `json:"shown"`
	Skipped int       `json:"skipped"`
}

// Discovery maps a DiscoveryView.
func Discovery(v usecase.DiscoveryView) DiscoveryDTO {
	dto := DiscoveryDTO{Available: v.Available, Enabled: v.Settings.Enabled, IntervalMinutes: v.Settings.IntervalMinutes, Rule: string(v.Settings.Rule)}
	if !v.Last.At.IsZero() {
		l := SyncResultDTO(v.Last)
		dto.LastSync = &l
	}
	return dto
}
