// Package usecase contains the calendar use cases: settings management and
// cached upcoming events.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
)

// Bounds for the number of days requested by the widget.
const (
	DefaultDays = 7
	MaxDays     = 31
)

// SettingsView never contains the password.
type SettingsView struct {
	Configured   bool
	CalendarURLs []string
	Username     string
	HasPassword  bool
}

// SettingsInput is the input of SaveSettings; an empty password keeps the stored one.
type SettingsInput struct {
	CalendarURLs []string
	Username     string
	Password     string
}

// EventView is an event at the use case boundary.
type EventView struct {
	UID, Summary, Location, Calendar string
	Start, End                       time.Time
	AllDay                           bool
}

// AgendaResponse lists upcoming events.
type AgendaResponse struct {
	Events    []EventView
	UpdatedAt time.Time
	Error     string
}

// Calendar is the input port of the module.
type Calendar interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Upcoming(ctx context.Context, days int) (AgendaResponse, error)
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port); Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// EventSource loads events of one calendar in a time range (output port).
type EventSource interface {
	Events(ctx context.Context, s domain.Settings, calendarURL string, from, to time.Time) ([]domain.Event, error)
}

type calendar struct {
	repo   SettingsRepository
	source EventSource
	clock  kernel.Clock
	mu     sync.RWMutex
	cached *AgendaResponse
}

// NewCalendar returns the calendar use cases.
func NewCalendar(repo SettingsRepository, source EventSource, clock kernel.Clock) Calendar {
	return &calendar{repo: repo, source: source, clock: clock}
}

// Settings implements Calendar.
func (uc *calendar) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{CalendarURLs: []string{}}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, CalendarURLs: s.CalendarURLs, Username: s.Username, HasPassword: true}, nil
}

// SaveSettings implements Calendar; every calendar is queried once so wrong
// URLs or credentials are reported immediately.
func (uc *calendar) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	password := in.Password
	if password == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			password = old.Password
		}
	}
	s, err := domain.NewSettings(in.CalendarURLs, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	agenda, err := uc.fetch(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.NewPublicError(kernel.ErrInvalid, "Could not load the calendar: "+err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(agenda)
	return SettingsView{Configured: true, CalendarURLs: s.CalendarURLs, Username: s.Username, HasPassword: true}, nil
}

// Upcoming implements Calendar (cache-aside).
func (uc *calendar) Upcoming(ctx context.Context, days int) (AgendaResponse, error) {
	if days <= 0 {
		days = DefaultDays
	}
	days = min(days, MaxDays)
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()
	if cached == nil {
		if _, err := uc.repo.Load(ctx); err != nil {
			return AgendaResponse{}, err
		}
		_ = uc.Refresh(ctx) // errors are reported through AgendaResponse.Error
		uc.mu.RLock()
		cached = uc.cached
		uc.mu.RUnlock()
		if cached == nil {
			return AgendaResponse{}, kernel.NewPublicError(kernel.ErrUnavailable, "The calendar server is not reachable.")
		}
	}
	now := uc.clock.Now()
	until := now.Add(time.Duration(days) * 24 * time.Hour)
	res := AgendaResponse{UpdatedAt: cached.UpdatedAt, Error: cached.Error, Events: []EventView{}}
	for _, e := range cached.Events {
		if e.Start.Before(until) && (domain.Event{End: e.End, AllDay: e.AllDay}).EndsAfter(now) {
			res.Events = append(res.Events, e)
		}
	}
	return res, nil
}

// Refresh implements Calendar; it is run by the background job.
func (uc *calendar) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	agenda, err := uc.fetch(ctx, s)
	if err != nil {
		uc.mu.Lock()
		res := AgendaResponse{}
		if uc.cached != nil {
			res = *uc.cached
		}
		res.Error = "The calendar server is not reachable; showing the last known events."
		uc.cached = &res
		uc.mu.Unlock()
		return err
	}
	uc.store(agenda)
	return nil
}

func (uc *calendar) fetch(ctx context.Context, s domain.Settings) (AgendaResponse, error) {
	now := uc.clock.Now()
	from := now.Add(-24 * time.Hour) // include today's all-day events in every time zone
	to := now.Add(domain.Horizon)
	var all []domain.Event
	var failed []string
	for _, u := range s.CalendarURLs {
		events, err := uc.source.Events(ctx, s, u, from, to)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", domain.CalendarName(u), err))
			continue
		}
		all = append(all, events...)
	}
	if len(failed) == len(s.CalendarURLs) {
		return AgendaResponse{}, &fetchError{failures: failed}
	}
	domain.SortEvents(all)
	res := AgendaResponse{UpdatedAt: now, Events: make([]EventView, 0, len(all))}
	if len(failed) > 0 {
		res.Error = "Some calendars could not be loaded: " + strings.Join(failed, "; ")
	}
	for _, e := range all {
		res.Events = append(res.Events, EventView{UID: e.UID, Summary: e.Summary, Location: e.Location, Calendar: e.Calendar, Start: e.Start, End: e.End, AllDay: e.AllDay})
	}
	return res, nil
}

// fetchError lists the calendars that could not be loaded. Its message is
// shown to the user, so it contains no error-kind suffix.
type fetchError struct{ failures []string }

func (e *fetchError) Error() string { return strings.Join(e.failures, "; ") }

// Unwrap classifies the error as an unavailable dependency.
func (e *fetchError) Unwrap() error { return kernel.ErrUnavailable }

func (uc *calendar) store(a AgendaResponse) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	a.Events = slices.Clone(a.Events)
	uc.cached = &a
}
