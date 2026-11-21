// Package presenter maps backup files to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/backup/usecase"
)

// FileDTO is a backup in API responses.
type FileDTO struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// File maps a FileView.
func File(v usecase.FileView) FileDTO { return FileDTO(v) }

// Files maps file views.
func Files(vs []usecase.FileView) []FileDTO {
	out := make([]FileDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, File(v))
	}
	return out
}

// StatusDTO is the traffic light of GET /api/backups/status (#677). Times
// are null when there was none; the errors are the server's own words.
type StatusDTO struct {
	Health          string     `json:"health"`
	LastBackupAt    *time.Time `json:"lastBackupAt"`
	Count           int        `json:"count"`
	IntervalHours   float64    `json:"intervalHours"`
	StaleAfterHours float64    `json:"staleAfterHours"`
	FailedAt        *time.Time `json:"failedAt"`
	Error           string     `json:"error"`
	OffsiteFailedAt *time.Time `json:"offsiteFailedAt"`
	OffsiteError    string     `json:"offsiteError"`
}

// Status maps a status view.
func Status(v usecase.StatusView) StatusDTO {
	return StatusDTO{
		Health:          string(v.Health),
		LastBackupAt:    optional(v.LastBackupAt),
		Count:           v.Count,
		IntervalHours:   v.Interval.Hours(),
		StaleAfterHours: v.StaleAfter.Hours(),
		FailedAt:        optional(v.FailedAt),
		Error:           v.Error,
		OffsiteFailedAt: optional(v.OffsiteFailedAt),
		OffsiteError:    v.OffsiteError,
	}
}

func optional(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
