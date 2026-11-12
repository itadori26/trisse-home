// Package presenter maps speedtest results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"
	"time"

	"trisse-home/internal/modules/speedtest/domain"
	"trisse-home/internal/modules/speedtest/usecase"
)

// SettingsDTO is returned by GET/PUT /api/speedtest/settings; the token
// is never returned.
type SettingsDTO struct {
	Configured  bool    `json:"configured"`
	URL         string  `json:"url"`
	HasToken    bool    `json:"hasToken"`
	MinDownMbps float64 `json:"minDownMbps"`
	MinUpMbps   float64 `json:"minUpMbps"`
}

// MeasurementDTO is one result in Mbit/s and milliseconds.
type MeasurementDTO struct {
	At           time.Time `json:"at"`
	DownloadMbps float64   `json:"downloadMbps"`
	UploadMbps   float64   `json:"uploadMbps"`
	PingMs       float64   `json:"pingMs"`
}

// ResultsDTO is returned by GET /api/speedtest/results.
type ResultsDTO struct {
	Latest    *MeasurementDTO  `json:"latest"`
	History   []MeasurementDTO `json:"history"`
	UpdatedAt time.Time        `json:"updatedAt"`
	Error     string           `json:"error"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func measurement(m domain.Measurement) MeasurementDTO {
	return MeasurementDTO{At: m.At, DownloadMbps: round1(domain.Mbps(m.DownloadBps)), UploadMbps: round1(domain.Mbps(m.UploadBps)), PingMs: round1(m.PingMs)}
}

// Results maps a ResultsView.
func Results(v usecase.ResultsView) ResultsDTO {
	dto := ResultsDTO{History: make([]MeasurementDTO, 0, len(v.History)), UpdatedAt: v.UpdatedAt, Error: v.Error}
	if v.Latest != nil {
		m := measurement(*v.Latest)
		dto.Latest = &m
	}
	for _, m := range v.History {
		dto.History = append(dto.History, measurement(m))
	}
	return dto
}
