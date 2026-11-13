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
	// Description is plain text; URL an http(s) link or empty (#163).
	Description string `json:"description"`
	URL         string `json:"url"`
	Calendar    string `json:"calendar"`
	// Color is the subscription colour (#rrggbb), empty for CalDAV calendars.
	Color  string `json:"color"`
	Start  string `json:"start"`
	End    string `json:"end"`
	AllDay bool   `json:"allDay"`
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
	ev := EventDTO{UID: e.UID, Summary: e.Summary, Location: e.Location, Description: e.Description, URL: e.URL,
		Calendar: e.Calendar, Color: e.Color, AllDay: e.AllDay}
	if e.AllDay {
		ev.Start, ev.End = e.Start.Format(time.DateOnly), e.End.Format(time.DateOnly)
	} else {
		ev.Start, ev.End = e.Start.Format(time.RFC3339), e.End.Format(time.RFC3339)
	}
	return ev
}

// SubscriptionDTO is an ICS subscription; the full URL is never returned
// because it may contain a private token.
type SubscriptionDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	URLHint        string `json:"urlHint"`
	Color          string `json:"color"`
	RefreshMinutes int    `json:"refreshMinutes"`
}

// Subscription maps a SubscriptionView.
func Subscription(v usecase.SubscriptionView) SubscriptionDTO { return SubscriptionDTO(v) }

// Subscriptions maps subscription views.
func Subscriptions(list []usecase.SubscriptionView) []SubscriptionDTO {
	out := make([]SubscriptionDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Subscription(v))
	}
	return out
}
