package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/services/domain"
)

// UptimeStore keeps daily aggregates and incidents (output port, #81).
type UptimeStore interface {
	AddCheck(ctx context.Context, serviceID, day string, online, maintenance bool) error
	// Aggregates returns all aggregates from sinceDay on.
	Aggregates(ctx context.Context, sinceDay string) ([]domain.DailyAggregate, error)
	PruneAggregatesBefore(ctx context.Context, day string) error
	OpenIncident(ctx context.Context, serviceID string) (domain.Incident, bool, error)
	SaveIncident(ctx context.Context, i domain.Incident) error
	// Incidents returns the newest incidents first; empty serviceID = all.
	Incidents(ctx context.Context, serviceID string, limit int) ([]domain.Incident, error)
	PruneIncidentsBefore(ctx context.Context, t time.Time) error
}

// MaintenanceReader tells whether a service is in a maintenance window
// (output port, implemented by the notifications module, #69).
type MaintenanceReader interface {
	UnderMaintenance(ctx context.Context, kind, key, label string, at time.Time) bool
}

// UptimeView summarises one service.
type UptimeView struct {
	ServiceID string
	// Percent values; nil when there is no data for the window.
	Day, Week, Month, Quarter *float64
	// Days are the last 90 days, oldest first; nil = no data.
	Days []*float64
}

func pct(v float64, ok bool) *float64 {
	if !ok {
		return nil
	}
	return &v
}

// recordUptime updates the aggregates and opens or closes incidents.
func (uc *services) recordUptime(ctx context.Context, s domain.Service, online bool, cause string, now time.Time, maintenance bool) {
	if uc.d.Uptime == nil {
		return
	}
	_ = uc.d.Uptime.AddCheck(ctx, s.ID, domain.DayOf(now.In(uc.d.Location)), online, maintenance) // best effort
	uc.mu.Lock()
	tr, known := uc.tracks[s.ID]
	uc.mu.Unlock()
	if !known {
		// After a restart an open incident is the baseline "offline".
		_, open, err := uc.d.Uptime.OpenIncident(ctx, s.ID)
		if err != nil {
			return
		}
		tr = domain.IncidentTrack{Known: true, Online: !open}
	}
	if maintenance && !online {
		// Planned downtime opens no incident.
		uc.mu.Lock()
		uc.tracks[s.ID] = tr
		uc.mu.Unlock()
		return
	}
	tr, tn := tr.Advance(online, now, cause)
	uc.mu.Lock()
	uc.tracks[s.ID] = tr
	uc.mu.Unlock()
	if !tn.Changed {
		return
	}
	if !tn.Online {
		_ = uc.d.Uptime.SaveIncident(ctx, domain.Incident{ID: uc.d.IDs.NewID(), ServiceID: s.ID, Start: tn.At, Cause: tn.Cause})
		return
	}
	if in, open, err := uc.d.Uptime.OpenIncident(ctx, s.ID); err == nil && open {
		in.End = tn.At
		_ = uc.d.Uptime.SaveIncident(ctx, in)
	}
}

// pruneUptime drops data older than the retention.
func (uc *services) pruneUptime(ctx context.Context, now time.Time) error {
	if uc.d.Uptime == nil {
		return nil
	}
	cutoff := now.Add(-domain.AggregateRetention)
	if err := uc.d.Uptime.PruneAggregatesBefore(ctx, domain.DayOf(cutoff.In(uc.d.Location))); err != nil {
		return fmt.Errorf("prune aggregates: %w", err)
	}
	return uc.d.Uptime.PruneIncidentsBefore(ctx, cutoff)
}

// Uptime implements Services.
func (uc *services) Uptime(ctx context.Context) ([]UptimeView, error) {
	if uc.d.Uptime == nil {
		return []UptimeView{}, nil
	}
	now := uc.d.Clock.Now().In(uc.d.Location)
	const days = 90
	first := now.AddDate(0, 0, -(days - 1))
	aggs, err := uc.d.Uptime.Aggregates(ctx, domain.DayOf(first))
	if err != nil {
		return nil, fmt.Errorf("load aggregates: %w", err)
	}
	raw, err := uc.d.History.Since(ctx, now.Add(-24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	byService := map[string]map[string]domain.DailyAggregate{}
	for _, a := range aggs {
		if byService[a.ServiceID] == nil {
			byService[a.ServiceID] = map[string]domain.DailyAggregate{}
		}
		byService[a.ServiceID][a.Day] = a
	}
	out := make([]UptimeView, 0, len(list))
	for _, s := range list {
		v := UptimeView{ServiceID: s.ID, Days: make([]*float64, 0, days)}
		window := func(n int) *float64 {
			var sel []domain.DailyAggregate
			for i := 0; i < n; i++ {
				if a, ok := byService[s.ID][domain.DayOf(now.AddDate(0, 0, -i))]; ok {
					sel = append(sel, a)
				}
			}
			return pct(domain.Uptime(sel))
		}
		v.Week, v.Month, v.Quarter = window(7), window(30), window(90)
		for i := days - 1; i >= 0; i-- {
			a, ok := byService[s.ID][domain.DayOf(now.AddDate(0, 0, -i))]
			if !ok {
				v.Days = append(v.Days, nil)
				continue
			}
			v.Days = append(v.Days, pct(domain.Uptime([]domain.DailyAggregate{a})))
		}
		var day domain.DailyAggregate
		for _, p := range raw {
			if p.ServiceID != s.ID {
				continue
			}
			day.Checks++
			switch {
			case p.Maintenance:
				day.Maintenance++
			case p.Online:
				day.Online++
			}
		}
		v.Day = pct(domain.Uptime([]domain.DailyAggregate{day}))
		out = append(out, v)
	}
	return out, nil
}

// Incidents implements Services.
func (uc *services) Incidents(ctx context.Context, serviceID string, limit int) ([]domain.Incident, error) {
	if uc.d.Uptime == nil {
		return []domain.Incident{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return uc.d.Uptime.Incidents(ctx, serviceID, limit)
}
