package presenter

import (
	"time"

	"trisse-home/internal/modules/calendar/usecase"
)

// SettingsDTO is returned by GET/PUT /api/calendar/settings; the password is never returned.
type SettingsDTO struct {
	Configured   bool     `json:"configured"`
	CalendarURLs []string `json:"calendarUrls"`
	Username     string   `json:"username"`
	HasPassword  bool     `json:"hasPassword"`
}

// EventDTO is one event. For all-day events Start and End are dates
// (YYYY-MM-DD, End exclusive); otherwise RFC 3339 timestamps.
type EventDTO struct {
	UID      string `json:"uid"`
	Summary  string `json:"summary"`
	Location string `json:"location"`
	Calendar string `json:"calendar"`
	Start    string `json:"start"`
	End      string `json:"end"`
	AllDay   bool   `json:"allDay"`
}

// AgendaDTO is returned by GET /api/calendar/events.
type AgendaDTO struct {
	Events    []EventDTO `json:"events"`
	UpdatedAt time.Time  `json:"updatedAt"`
	Error     string     `json:"error"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO {
	urls := v.CalendarURLs
	if urls == nil {
		urls = []string{}
	}
	return SettingsDTO{Configured: v.Configured, CalendarURLs: urls, Username: v.Username, HasPassword: v.HasPassword}
}

// Agenda maps an AgendaResponse.
func Agenda(r usecase.AgendaResponse) AgendaDTO {
	dto := AgendaDTO{UpdatedAt: r.UpdatedAt, Error: r.Error, Events: make([]EventDTO, 0, len(r.Events))}
	for _, e := range r.Events {
		dto.Events = append(dto.Events, Event(e))
	}
	return dto
}

// Event maps one EventView; all-day events use dates, timed events RFC 3339.
func Event(e usecase.EventView) EventDTO {
	ev := EventDTO{UID: e.UID, Summary: e.Summary, Location: e.Location, Calendar: e.Calendar, AllDay: e.AllDay}
	if e.AllDay {
		ev.Start, ev.End = e.Start.Format(time.DateOnly), e.End.Format(time.DateOnly)
	} else {
		ev.Start, ev.End = e.Start.Format(time.RFC3339), e.End.Format(time.RFC3339)
	}
	return ev
}
