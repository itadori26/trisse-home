package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/shared/kernel"
)

type memUptime struct {
	mu        sync.Mutex
	aggs      map[string]domain.DailyAggregate
	incidents []domain.Incident
}

func (m *memUptime) AddCheck(_ context.Context, id, day string, online, maint bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	a := m.aggs[id+day]
	a.ServiceID, a.Day = id, day
	a.Checks++
	if maint {
		a.Maintenance++
	} else if online {
		a.Online++
	}
	m.aggs[id+day] = a
	return nil
}
func (m *memUptime) Aggregates(context.Context, string) ([]domain.DailyAggregate, error) {
	var out []domain.DailyAggregate
	for _, a := range m.aggs {
		out = append(out, a)
	}
	return out, nil
}
func (m *memUptime) PruneAggregatesBefore(context.Context, string) error { return nil }
func (m *memUptime) OpenIncident(_ context.Context, id string) (domain.Incident, bool, error) {
	for _, i := range m.incidents {
		if i.ServiceID == id && i.Open() {
			return i, true, nil
		}
	}
	return domain.Incident{}, false, nil
}
func (m *memUptime) SaveIncident(_ context.Context, in domain.Incident) error {
	for k := range m.incidents {
		if m.incidents[k].ID == in.ID {
			m.incidents[k] = in
			return nil
		}
	}
	m.incidents = append(m.incidents, in)
	return nil
}
func (m *memUptime) Incidents(context.Context, string, int) ([]domain.Incident, error) {
	return m.incidents, nil
}
func (m *memUptime) PruneIncidentsBefore(context.Context, time.Time) error { return nil }

type tickClock struct{ t time.Time }

func (c *tickClock) Now() time.Time { return c.t }

type window struct{ from, to time.Time }

func (w window) UnderMaintenance(_ context.Context, kind, _, label string, at time.Time) bool {
	return kind == "service" && label == "NAS" && !at.Before(w.from) && at.Before(w.to)
}

func TestUptimeAndIncidents(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	clock := &tickClock{t: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)}
	checker := fakeChecker{down: map[string]bool{}}
	store := &memUptime{aggs: map[string]domain.DailyAggregate{}}
	maint := window{from: clock.t.Add(10 * time.Minute), to: clock.t.Add(20 * time.Minute)}
	uc := NewServices(Deps{Repo: repo, Status: &memStatus{m: map[string]domain.Status{}}, Checker: checker, History: &memHistory{},
		Observations: kernel.DiscardObservations{}, Clock: clock, IDs: &ids{}, Uptime: store, Maintenance: maint})
	nas, err := uc.Create(ctx, ServiceInput{Name: "NAS", URL: "http://nas.lan"})
	if err != nil {
		t.Fatal(err)
	}
	run := func(down bool, minutes int) {
		for range minutes {
			checker.down["http://nas.lan"] = down
			_ = uc.CheckAll(ctx)
			clock.t = clock.t.Add(time.Minute)
		}
	}
	run(false, 2) // 10:00-10:01 online
	run(true, 3)  // 10:02-10:04 outage -> incident from 10:02
	run(false, 5) // recovered at 10:05
	run(true, 5)  // 10:10-10:14 planned maintenance: no incident
	run(false, 2)
	if len(store.incidents) != 1 {
		t.Fatalf("incidents: %+v", store.incidents)
	}
	in := store.incidents[0]
	if !in.Start.Equal(time.Date(2026, 11, 12, 10, 2, 0, 0, time.UTC)) || in.Open() || in.Duration(clock.t) != 3*time.Minute || in.Cause == "" {
		t.Fatalf("incident %+v", in)
	}
	views, err := uc.Uptime(ctx)
	if err != nil || len(views) != 1 || views[0].ServiceID != nas.ID {
		t.Fatalf("%+v %v", views, err)
	}
	// Create checks once, then 17 checks: 18 in total, 7 inside the
	// maintenance window (10:10-10:20) are not counted, 3 were offline:
	// 8 of the 11 counted checks were online.
	v := views[0]
	want := 100 * 8.0 / 11.0
	if v.Month == nil || *v.Month != want || v.Day == nil || *v.Day != want || len(v.Days) != 90 || v.Days[89] == nil {
		t.Fatalf("uptime month %v day %v", deref(v.Month), deref(v.Day))
	}
}

func deref(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}
