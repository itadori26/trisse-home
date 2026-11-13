package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/briefing"
	"trisse-home/internal/modules/calendar"
	"trisse-home/internal/modules/services"
	"trisse-home/internal/modules/timeline"
	"trisse-home/internal/modules/todos"
	"trisse-home/internal/modules/waste"
	"trisse-home/internal/modules/weather"
	"trisse-home/internal/shared/kernel"
)

// This file is the composition root of the briefing (#52, ADR 0016): the
// briefing owns its read ports, and these small adapters implement them on
// top of the source modules' public read methods. Modules never import
// each other; a disabled source module simply leaves its port nil.

// briefingSources finds the enabled source modules in the registry.
func briefingSources(registry *modules.Registry) briefing.Sources {
	var src briefing.Sources
	for _, m := range registry.All() {
		switch m := m.(type) {
		case *todos.Module:
			src.Todos = todoReader{m}
		case *calendar.Module:
			src.Events = eventReader{m}
		case *weather.Module:
			src.Warnings = warningReader{m}
		case *waste.Module:
			src.Collection = collectionReader{m}
		case *services.Module:
			src.Services = serviceReader{m}
		case *timeline.Module:
			src.Incidents = incidentReader{m}
		}
	}
	return src
}

// notConfigured reports whether a source has simply not been set up yet
// (no waste calendar, no CalDAV account). That is not an outage: the
// section stays empty instead of being reported as unavailable.
func notConfigured(err error) bool { return errors.Is(err, kernel.ErrNotFound) }

type todoReader struct{ m *todos.Module }

// DueTodos implements the briefing's DueTodosReader.
func (r todoReader) DueTodos(ctx context.Context, until time.Time) ([]briefing.TodoFact, error) {
	list, err := r.m.OpenTodos(ctx)
	if err != nil {
		return nil, err
	}
	var out []briefing.TodoFact
	for _, t := range list {
		if !t.Due.IsZero() && !t.Due.After(until) {
			out = append(out, briefing.TodoFact{Title: t.Title, Due: t.Due, Overdue: t.Overdue})
		}
	}
	return out, nil
}

type eventReader struct{ m *calendar.Module }

// UpcomingEvents implements the briefing's UpcomingEventsReader.
func (r eventReader) UpcomingEvents(ctx context.Context, until time.Time) ([]briefing.EventFact, error) {
	agenda, err := r.m.Upcoming(ctx, 2)
	if notConfigured(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if agenda.Error != "" && len(agenda.Events) == 0 {
		return nil, errors.New(agenda.Error)
	}
	var out []briefing.EventFact
	for _, e := range agenda.Events {
		if e.Start.Before(until) {
			out = append(out, briefing.EventFact{Title: e.Summary, Location: e.Location, Start: e.Start, AllDay: e.AllDay})
		}
	}
	return out, nil
}

type warningReader struct{ m *weather.Module }

// WeatherWarnings implements the briefing's WeatherWarningReader.
func (r warningReader) WeatherWarnings(ctx context.Context, lat, lon float64) ([]briefing.WarningFact, error) {
	ws, err := r.m.Warnings(ctx, lat, lon)
	if err != nil {
		return nil, err
	}
	levels := map[string]int{"minor": 1, "moderate": 2, "severe": 3, "extreme": 4}
	out := make([]briefing.WarningFact, 0, len(ws))
	for _, w := range ws {
		headline := w.Headline
		if headline == "" {
			headline = w.Event
		}
		out = append(out, briefing.WarningFact{Headline: headline, Level: levels[string(w.Severity)], Expires: w.Expires})
	}
	return out, nil
}

type collectionReader struct{ m *waste.Module }

// Collections implements the briefing's CollectionReader.
func (r collectionReader) Collections(ctx context.Context, until time.Time) ([]briefing.CollectionFact, error) {
	o, err := r.m.Upcoming(ctx)
	if notConfigured(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []briefing.CollectionFact
	for _, u := range o.Types {
		for _, d := range u.Dates {
			// Dates are calendar days; compare them as such.
			day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, until.Location())
			if day.Before(until) {
				out = append(out, briefing.CollectionFact{Type: u.Style.Label, Day: time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)})
			}
		}
	}
	return out, nil
}

type serviceReader struct{ m *services.Module }

// DownServices implements the briefing's ServiceStatusReader.
func (r serviceReader) DownServices(ctx context.Context) ([]briefing.ServiceFact, error) {
	list, err := r.m.Services(ctx)
	if err != nil {
		return nil, err
	}
	var out []briefing.ServiceFact
	for _, s := range list {
		if s.State == "offline" {
			out = append(out, briefing.ServiceFact{Name: s.Name, Error: s.Error})
		}
	}
	return out, nil
}

type incidentReader struct{ m *timeline.Module }

// maxIncidentScan bounds how many recent timeline events are looked at.
const maxIncidentScan = 200

// Incidents implements the briefing's IncidentReader: warnings and
// failures since the given time; recoveries ("… is OK again") are info
// events and therefore left out.
func (r incidentReader) Incidents(ctx context.Context, since time.Time) ([]briefing.IncidentFact, error) {
	events, err := r.m.Recent(ctx, maxIncidentScan)
	if err != nil {
		return nil, err
	}
	var out []briefing.IncidentFact
	for _, e := range events {
		if e.At.Before(since) || strings.EqualFold(string(e.Severity), "info") {
			continue
		}
		out = append(out, briefing.IncidentFact{At: e.At, Message: e.Message, Critical: string(e.Severity) == "critical"})
	}
	return out, nil
}

// briefingPublisher delivers the daily briefing through the notifications
// module (#70) – deliberately not through the alert relay, so the daily
// summary does not show up in the timeline.
type briefingPublisher struct{ sink kernel.AlertSink }

// Publish implements briefing.Publisher.
func (p briefingPublisher) Publish(ctx context.Context, title, body string, urgent bool) {
	p.sink.Alert(ctx, kernel.Alert{Kind: "briefing", Title: title, Body: body, Urgent: urgent})
}
