// Package presenter maps the Nextcloud status to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/nextcloud/usecase"
)

// SettingsDTO is returned by the settings endpoints (never the password).
type SettingsDTO struct {
	Configured  bool   `json:"configured"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword"`
}

// QuotaDTO is the account's storage in bytes; total 0 means unlimited.
type QuotaDTO struct {
	Used    int64   `json:"used"`
	Free    int64   `json:"free"`
	Total   int64   `json:"total"`
	Percent float64 `json:"percent"`
}

// ServerDTO comes from the serverinfo app.
type ServerDTO struct {
	FreeSpace     int64 `json:"freeSpace"`
	Users         int   `json:"users"`
	ActiveUsers5m int   `json:"activeUsers5m"`
	ActiveUsers1d int   `json:"activeUsers1d"`
	Files         int   `json:"files"`
	AppUpdates    int   `json:"appUpdates"`
	DatabaseSize  int64 `json:"databaseSize"`
}

// StatusDTO is returned by GET /api/nextcloud/status.
type StatusDTO struct {
	Version     string     `json:"version"`
	Maintenance bool       `json:"maintenance"`
	Quota       QuotaDTO   `json:"quota"`
	Server      *ServerDTO `json:"server"`
	ServerHint  string     `json:"serverHint"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Error       string     `json:"error"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Status maps a StatusView.
func Status(v usecase.StatusView) StatusDTO {
	s := v.Status
	dto := StatusDTO{
		Version: s.Version, Maintenance: s.Maintenance, ServerHint: s.ServerHint, UpdatedAt: v.UpdatedAt, Error: v.Error,
		Quota: QuotaDTO{Used: s.Quota.Used, Free: s.Quota.Free, Total: s.Quota.Total, Percent: s.Quota.Percent()},
	}
	if s.Server != nil {
		srv := ServerDTO(*s.Server)
		dto.Server = &srv
	}
	return dto
}
