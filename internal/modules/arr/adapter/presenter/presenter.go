// Package presenter maps *arr use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"
	"time"

	"trisse-home/internal/modules/arr/domain"
	"trisse-home/internal/modules/arr/usecase"
)

// InstanceDTO is a configured instance; the API key is never returned.
type InstanceDTO struct {
	ID   string `json:"id"`
	App  string `json:"app" tstype:"'sonarr' | 'radarr' | 'prowlarr'"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ReleaseDTO is an upcoming episode or movie.
type ReleaseDTO struct {
	Instance string    `json:"instance"`
	Title    string    `json:"title"`
	Detail   string    `json:"detail"`
	Date     time.Time `json:"date"`
	HasFile  bool      `json:"hasFile"`
}

// QueueItemDTO is a download.
type QueueItemDTO struct {
	Instance        string  `json:"instance"`
	Title           string  `json:"title"`
	Status          string  `json:"status"`
	Progress        float64 `json:"progress"`
	TimeLeftSeconds int64   `json:"timeLeftSeconds"`
	Error           string  `json:"error"`
}

// HealthIssueDTO is a problem reported by an app.
type HealthIssueDTO struct {
	Instance string `json:"instance"`
	Severity string `json:"severity" tstype:"'notice' | 'warning' | 'error'"`
	Message  string `json:"message"`
}

// OverviewDTO is returned by GET /api/arr/overview.
type OverviewDTO struct {
	Upcoming []ReleaseDTO     `json:"upcoming"`
	Queue    []QueueItemDTO   `json:"queue"`
	Health   []HealthIssueDTO `json:"health"`
	Missing  map[string]int   `json:"missing"`
	Errors   []string         `json:"errors"`
}

// Instance maps an InstanceView.
func Instance(v usecase.InstanceView) InstanceDTO { return InstanceDTO(v) }

// Instances maps instance views.
func Instances(list []usecase.InstanceView) []InstanceDTO {
	out := make([]InstanceDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Instance(v))
	}
	return out
}

// Overview maps the merged overview.
func Overview(o domain.Overview) OverviewDTO {
	dto := OverviewDTO{Upcoming: make([]ReleaseDTO, 0, len(o.Upcoming)), Queue: make([]QueueItemDTO, 0, len(o.Queue)),
		Health: make([]HealthIssueDTO, 0, len(o.Health)), Missing: o.Missing, Errors: o.Errors}
	if dto.Missing == nil {
		dto.Missing = map[string]int{}
	}
	if dto.Errors == nil {
		dto.Errors = []string{}
	}
	for _, r := range o.Upcoming {
		dto.Upcoming = append(dto.Upcoming, ReleaseDTO(r))
	}
	for _, q := range o.Queue {
		dto.Queue = append(dto.Queue, QueueItemDTO{Instance: q.Instance, Title: q.Title, Status: q.Status,
			Progress: math.Round(q.Progress), TimeLeftSeconds: int64(q.TimeLeft.Seconds()), Error: q.Error})
	}
	for _, h := range o.Health {
		dto.Health = append(dto.Health, HealthIssueDTO{Instance: h.Instance, Severity: string(h.Severity), Message: h.Message})
	}
	return dto
}
