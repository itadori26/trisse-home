// Package presenter maps stops and departures to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/transit/domain"
)

// StopDTO is a stop found by the search.
type StopDTO struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// DepartureDTO is one departure; expected is null without real-time data.
type DepartureDTO struct {
	Line         string     `json:"line"`
	Headsign     string     `json:"headsign"`
	Mode         string     `json:"mode"`
	Planned      time.Time  `json:"planned"`
	Expected     *time.Time `json:"expected"`
	DelayMinutes int        `json:"delayMinutes"`
	Track        string     `json:"track"`
	Cancelled    bool       `json:"cancelled"`
}

// Stops maps stops.
func Stops(ss []domain.Stop) []StopDTO {
	out := make([]StopDTO, 0, len(ss))
	for _, s := range ss {
		out = append(out, StopDTO(s))
	}
	return out
}

// Departures maps departures.
func Departures(ds []domain.Departure) []DepartureDTO {
	out := make([]DepartureDTO, 0, len(ds))
	for _, d := range ds {
		dto := DepartureDTO{Line: d.Line, Headsign: d.Headsign, Mode: d.Mode, Planned: d.Planned, DelayMinutes: d.DelayMinutes(),
			Track: d.Track, Cancelled: d.Cancelled}
		if d.RealTime && !d.Expected.IsZero() {
			e := d.Expected
			dto.Expected = &e
		}
		out = append(out, dto)
	}
	return out
}
