// Package presenter maps backup sources to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/backupmonitor/usecase"
)

// SourceDTO is a backup source in API responses.
type SourceDTO struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	MaxAgeHours   float64    `json:"maxAgeHours"`
	Kind          string     `json:"kind" tstype:"'backup' | 'check'"`
	Cron          string     `json:"cron"`
	GraceMinutes  float64    `json:"graceMinutes"`
	LastStartAt   *time.Time `json:"lastStartAt"`
	LastDurationS float64    `json:"lastDurationSeconds"`
	State         string     `json:"state" tstype:"'ok' | 'overdue' | 'failed' | 'waiting' | 'running'"`
	LastSuccessAt *time.Time `json:"lastSuccessAt"`
	LastFailureAt *time.Time `json:"lastFailureAt"`
	LastMessage   string     `json:"lastMessage"`
}

// CreatedDTO is returned once after creating a source or rotating its token.
type CreatedDTO struct {
	Source  SourceDTO `json:"source"`
	PingURL string    `json:"pingUrl"`
}

func optional(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// Source maps a SourceView.
func Source(v usecase.SourceView) SourceDTO {
	return SourceDTO{ID: v.ID, Name: v.Name, MaxAgeHours: v.MaxAge.Hours(), State: v.State, Kind: v.Kind, Cron: v.Cron,
		GraceMinutes: v.Grace.Minutes(), LastStartAt: optional(v.LastStartAt), LastDurationS: v.LastDuration.Seconds(),
		LastSuccessAt: optional(v.LastSuccessAt), LastFailureAt: optional(v.LastFailureAt), LastMessage: v.LastMessage}
}

// Sources maps a list of SourceViews.
func Sources(vs []usecase.SourceView) []SourceDTO {
	out := make([]SourceDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Source(v))
	}
	return out
}

// Created maps a CreatedSource; baseURL is the externally visible origin.
func Created(c usecase.CreatedSource, baseURL string) CreatedDTO {
	return CreatedDTO{Source: Source(c.Source), PingURL: baseURL + "/api/backup-status/ping/" + c.Token}
}
