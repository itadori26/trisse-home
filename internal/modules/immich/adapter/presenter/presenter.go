// Package presenter maps Immich use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"

	"trisse-home/internal/modules/immich/domain"
	"trisse-home/internal/modules/immich/usecase"
)

// SettingsDTO is returned by GET/PUT /api/immich/settings; the API key is
// never returned.
type SettingsDTO struct {
	Configured bool   `json:"configured"`
	URL        string `json:"url"`
	HasAPIKey  bool   `json:"hasApiKey"`
}

// MemoryDTO is "on this day"; thumbnails are at /api/immich/thumbnails/{id}.
type MemoryDTO struct {
	YearsAgo int      `json:"yearsAgo"`
	AssetIDs []string `json:"assetIds"`
}

// OverviewDTO is returned by GET /api/immich/overview.
type OverviewDTO struct {
	Version     string      `json:"version"`
	Photos      int         `json:"photos"`
	Videos      int         `json:"videos"`
	DiskUsed    int64       `json:"diskUsed"`
	DiskTotal   int64       `json:"diskTotal"`
	DiskPercent float64     `json:"diskPercent"`
	Memories    []MemoryDTO `json:"memories"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Overview maps the overview.
func Overview(o domain.Overview) OverviewDTO {
	dto := OverviewDTO{Version: o.Version, Photos: o.Stats.Photos, Videos: o.Stats.Videos, DiskUsed: o.Stats.DiskUsed,
		DiskTotal: o.Stats.DiskTotal, DiskPercent: math.Round(o.Stats.DiskPercent()*10) / 10, Memories: make([]MemoryDTO, 0, len(o.Memories))}
	for _, m := range o.Memories {
		dto.Memories = append(dto.Memories, MemoryDTO(m))
	}
	return dto
}
