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
	UID, Summary, Location, Calendar, Color string
	// Description (plain text) and URL (http(s) only) for the detail view (#163).
	Description, URL string
	Start, End       time.Time
	AllDay           bool
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
	// Range returns the events overlapping [from, to) (at most 42 days).
	Range(ctx context.Context, from, to time.Time) (AgendaResponse, error)
	// CreateEvent writes a new event to one of the configured calendars.
	CreateEvent(ctx context.Context, in NewEventInput) (EventView, error)
	Refresh(ctx context.Context) error
}

// NewEventInput describes an event to create. For all-day events Start and
// End are dates and End is inclusive (a one-day event has End == Start).
type NewEventInput struct {
	CalendarURL, Summary, Location string
	AllDay                         bool
	Start, End                     time.Time
}

// EventWriter creates events in a calendar (output port).
type EventWriter interface {
	Create(ctx context.Context, s domain.Settings, calendarURL string, e domain.Event) error
}

// SettingsRepository persists the settings (output port); Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// EventSource loads events of one CalDAV calendar in a time range (output port).
type EventSource interface {
	Events(ctx context.Context, s domain.Settings, calendarURL string, from, to time.Time) ([]domain.Event, error)
}

// FeedSource loads the events of an ICS subscription in a time range,
// recurrences expanded (output port, #98). The adapter decides when to
// download the feed again (refresh interval, conditional requests).
type FeedSource interface {
	Events(ctx context.Context, s domain.Subscription, from, to time.Time) ([]domain.Event, error)
}

// SubscriptionRepository persists ICS subscriptions (output port).
type SubscriptionRepository interface {
	List(ctx context.Context) ([]domain.Subscription, error)
	Get(ctx context.Context, id string) (domain.Subscription, error)
	Save(ctx context.Context, s domain.Subscription) error
	Delete(ctx context.Context, id string) error
}

// Deps bundles the output ports of the calendar module.
type Deps struct {
	Repo          SettingsRepository
	Source        EventSource
	Writer        EventWriter
	Subscriptions SubscriptionRepository
	Feeds         FeedSource
	IDs           kernel.IDGenerator
	Clock         kernel.Clock
}

// Service bundles both input ports; one value implements them so CalDAV
// calendars and subscriptions share the agenda cache.
type Service interface {
	Calendar
	Subscriptions
}

type calendar struct {
	repo   SettingsRepository
	source EventSource
	writer EventWriter
	subs   SubscriptionRepository
	feeds  FeedSource
	ids    kernel.IDGenerator
	clock  kernel.Clock
	mu     sync.RWMutex
	cached *AgendaResponse
	// window is the period covered by cached.
	windowFrom, windowTo time.Time
}

// NewCalendar returns the calendar use cases.
func NewCalendar(d Deps) Service {
	return &calendar{repo: d.Repo, source: d.Source, writer: d.Writer, subs: d.Subscriptions, feeds: d.Feeds, ids: d.IDs, clock: d.Clock}
}

// sources returns the CalDAV settings (nil when not configured) and the
// subscriptions; domain.ErrNotConfigured when there is neither.
func (uc *calendar) sources(ctx context.Context) (*domain.Settings, []domain.Subscription, error) {
	var caldav *domain.Settings
	s, err := uc.repo.Load(ctx)
	switch {
	case err == nil:
		caldav = &s
	case !errors.Is(err, domain.ErrNotConfigured):
		return nil, nil, fmt.Errorf("load settings: %w", err)
	}
	subs, err := uc.subs.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list subscriptions: %w", err)
	}
	if caldav == nil && len(subs) == 0 {
		return nil, nil, domain.ErrNotConfigured
	}
	return caldav, subs, nil
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
			wanted := []string{in.Username}
			for _, u := range in.CalendarURLs {
				if strings.TrimSpace(u) != "" {
					wanted = append(wanted, u)
				}
			}
			if kernel.TargetChanged(append([]string{old.Username}, old.CalendarURLs...), wanted) {
				return SettingsView{}, kernel.SecretRequired("password")
			}
			password = old.Password
		}
	}
	s, err := domain.NewSettings(in.CalendarURLs, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	now := uc.clock.Now()
	if _, err := uc.fetchRange(ctx, &s, nil, now, now.Add(domain.Horizon)); err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	_ = uc.Refresh(ctx) // errors are reported through AgendaResponse.Error
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
		if _, _, err := uc.sources(ctx); err != nil {
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

// Range implements Calendar. Periods inside the cached window are answered
// from the cache, others are loaded directly (and not cached).
func (uc *calendar) Range(ctx context.Context, from, to time.Time) (AgendaResponse, error) {
	if err := domain.CheckRange(from, to); err != nil {
		return AgendaResponse{}, err
	}
	caldav, subs, err := uc.sources(ctx)
	if err != nil {
		return AgendaResponse{}, err
	}
	uc.mu.RLock()
	cached, wFrom, wTo := uc.cached, uc.windowFrom, uc.windowTo
	uc.mu.RUnlock()
	src := cached
	if cached == nil || from.Before(wFrom) || to.After(wTo) {
		live, err := uc.fetchRange(ctx, caldav, subs, from, to)
		if err != nil {
			return AgendaResponse{}, err
		}
		src = &live
	}
	res := AgendaResponse{UpdatedAt: src.UpdatedAt, Error: src.Error, Events: []EventView{}}
	for _, e := range src.Events {
		if e.Start.Before(to) && e.End.After(from) {
			res.Events = append(res.Events, e)
		}
	}
	return res, nil
}

// CreateEvent implements Calendar.
func (uc *calendar) CreateEvent(ctx context.Context, in NewEventInput) (EventView, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return EventView{}, err
	}
	if !s.Contains(in.CalendarURL) {
		return EventView{}, domain.ErrUnknownCalendar
	}
	end := in.End
	if in.AllDay {
		end = end.AddDate(0, 0, 1) // inclusive date → exclusive end
	}
	e, err := domain.NewEvent(uc.ids.NewID()+"@trisse-home", in.Summary, in.Location, in.Start, end, in.AllDay)
	if err != nil {
		return EventView{}, err
	}
	if err := uc.writer.Create(ctx, s, in.CalendarURL, e); err != nil {
		return EventView{}, kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	_ = uc.Refresh(ctx) // the new event shows up in the agenda right away
	return EventView{UID: e.UID, Summary: e.Summary, Location: e.Location, Calendar: domain.CalendarName(in.CalendarURL),
		Start: e.Start, End: e.End, AllDay: e.AllDay}, nil
}

// Refresh implements Calendar; it is run by the background job.
func (uc *calendar) Refresh(ctx context.Context) error {
	caldav, subs, err := uc.sources(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		uc.mu.Lock()
		uc.cached = nil // the last source was removed
		uc.mu.Unlock()
		return nil
	}
	if err != nil {
		return err
	}
	agenda, err := uc.fetch(ctx, caldav, subs)
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

func (uc *calendar) fetch(ctx context.Context, caldav *domain.Settings, subs []domain.Subscription) (AgendaResponse, error) {
	now := uc.clock.Now()
	from := now.Add(-24 * time.Hour) // include today's all-day events in every time zone
	res, err := uc.fetchRange(ctx, caldav, subs, from, now.Add(domain.Horizon))
	if err == nil {
		uc.mu.Lock()
		uc.windowFrom, uc.windowTo = from, now.Add(domain.Horizon)
		uc.mu.Unlock()
	}
	return res, err
}

// fetchRange merges the events of every CalDAV calendar and subscription;
// it fails only when every source failed.
func (uc *calendar) fetchRange(ctx context.Context, caldav *domain.Settings, subs []domain.Subscription, from, to time.Time) (AgendaResponse, error) {
	now := uc.clock.Now()
	var all []domain.Event
	var failed []string
	total := len(subs)
	if caldav != nil {
		total += len(caldav.CalendarURLs)
		for _, u := range caldav.CalendarURLs {
			events, err := uc.source.Events(ctx, *caldav, u, from, to)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s (%v)", domain.CalendarName(u), err))
				continue
			}
			all = append(all, events...)
		}
	}
	for _, sub := range subs {
		events, err := uc.feeds.Events(ctx, sub, from, to)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", sub.Name, err))
			continue
		}
		for _, e := range events {
			e.Calendar, e.Color = sub.Name, sub.Color
			all = append(all, e)
		}
	}
	if total > 0 && len(failed) == total {
		return AgendaResponse{}, &fetchError{failures: failed}
	}
	domain.SortEvents(all)
	res := AgendaResponse{UpdatedAt: now, Events: make([]EventView, 0, len(all))}
	if len(failed) > 0 {
		res.Error = "Some calendars could not be loaded: " + strings.Join(failed, "; ")
	}
	for _, e := range all {
		res.Events = append(res.Events, EventView{UID: e.UID, Summary: e.Summary, Location: e.Location, Description: e.Description, URL: e.URL, Calendar: e.Calendar,
			Color: e.Color, Start: e.Start, End: e.End, AllDay: e.AllDay})
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
