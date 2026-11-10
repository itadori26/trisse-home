// Package presenter maps AdGuard responses to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"
	"time"

	"trisse-home/internal/modules/adguard/usecase"
)

// SettingsDTO is returned by GET/PUT /api/adguard/settings. The password is
// never returned.
type SettingsDTO struct {
	Configured  bool   `json:"configured"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword"`
}

// BlockedDomainDTO is a blocked domain with its count.
type BlockedDomainDTO struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

// StatsDTO is returned by GET /api/adguard/stats.
type StatsDTO struct {
	Queries           int                `json:"queries"`
	Blocked           int                `json:"blocked"`
	BlockedPercent    float64            `json:"blockedPercent"`
	AvgProcessingMs   float64            `json:"avgProcessingMs"`
	ProtectionEnabled bool               `json:"protectionEnabled"`
	TopBlocked        []BlockedDomainDTO `json:"topBlocked"`
	UpdatedAt         time.Time          `json:"updatedAt"`
	Error             string             `json:"error"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Stats maps a StatsView.
func Stats(v usecase.StatsView) StatsDTO {
	dto := StatsDTO{
		Queries: v.Queries, Blocked: v.Blocked, BlockedPercent: math.Round(v.BlockedPercent*10) / 10,
		AvgProcessingMs: math.Round(v.AvgProcessingMs*100) / 100, ProtectionEnabled: v.ProtectionEnabled,
		UpdatedAt: v.UpdatedAt, Error: v.Error, TopBlocked: make([]BlockedDomainDTO, 0, len(v.TopBlocked)),
	}
	for _, d := range v.TopBlocked {
		dto.TopBlocked = append(dto.TopBlocked, BlockedDomainDTO(d))
	}
	return dto
}
