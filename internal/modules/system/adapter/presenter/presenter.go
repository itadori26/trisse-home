package presenter

import (
	"math"
	"time"

	"trisse-home/internal/modules/system/usecase"
)

// UsageDTO is a used/total pair in bytes with the percentage.
type UsageDTO struct {
	Total   uint64  `json:"total"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

// DiskDTO is the usage of one file system.
type DiskDTO struct {
	Label string   `json:"label"`
	Usage UsageDTO `json:"usage"`
}

// SnapshotDTO is returned by GET /api/system.
type SnapshotDTO struct {
	CPUPercent    float64    `json:"cpuPercent"`
	CPUCount      int        `json:"cpuCount"`
	Load          [3]float64 `json:"load"`
	Memory        UsageDTO   `json:"memory"`
	Swap          UsageDTO   `json:"swap"`
	Disks         []DiskDTO  `json:"disks"`
	UptimeSeconds int64      `json:"uptimeSeconds"`
	Temperature   float64    `json:"temperature"`
	TakenAt       time.Time  `json:"takenAt"`
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func usage(u usecase.UsageView) UsageDTO {
	return UsageDTO{Total: u.Total, Used: u.Used, Percent: round1(u.Percent)}
}

// Snapshot maps a SnapshotView.
func Snapshot(s usecase.SnapshotView) SnapshotDTO {
	dto := SnapshotDTO{
		CPUPercent: round1(s.CPUPercent), CPUCount: s.CPUCount,
		Load:   [3]float64{s.Load1, s.Load5, s.Load15},
		Memory: usage(s.Memory), Swap: usage(s.Swap), Disks: make([]DiskDTO, 0, len(s.Disks)),
		UptimeSeconds: int64(s.Uptime.Seconds()), Temperature: round1(s.Temperature), TakenAt: s.TakenAt,
	}
	for _, d := range s.Disks {
		dto.Disks = append(dto.Disks, DiskDTO{Label: d.Label, Usage: usage(d.Usage)})
	}
	return dto
}
