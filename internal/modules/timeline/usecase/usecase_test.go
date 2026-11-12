package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/timeline/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	events []domain.Event
	days   int
}

func (m *memRepo) Add(_ context.Context, e domain.Event) error {
	m.events = append(m.events, e)
	return nil
}
func (m *memRepo) List(_ context.Context, f Filter) ([]domain.Event, error) {
	var out []domain.Event
	for i := len(m.events) - 1; i >= 0; i-- {
		e := m.events[i]
		if (f.Source == "" || e.Source == f.Source) && (f.Severity == "" || e.Severity == f.Severity) {
			out = append(out, e)
		}
	}
	return out, nil
}
func (m *memRepo) DeleteBefore(_ context.Context, t time.Time) error {
	var keep []domain.Event
	for _, e := range m.events {
		if !e.At.Before(t) {
			keep = append(keep, e)
		}
	}
	m.events = keep
	return nil
}
func (m *memRepo) RetentionDays(context.Context) (int, error)      { return m.days, nil }
func (m *memRepo) SetRetentionDays(_ context.Context, d int) error { m.days = d; return nil }

type clock struct{ t time.Time }

func (c *clock) Now() time.Time { return c.t }

type seq struct{ n int }

func (s *seq) NewID() string { s.n++; return fmt.Sprint(s.n) }

func TestTimeline(t *testing.T) {
	ctx := context.Background()
	repo, clk := &memRepo{}, &clock{t: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)}
	uc := NewTimeline(repo, clk, &seq{})
	nas := kernel.Observation{Kind: "service", Key: "s1", Label: "NAS", Healthy: true}
	for _, healthy := range []bool{true, false, false, false, true, true} {
		nas.Healthy = healthy
		uc.Observe(ctx, nas)
	}
	uc.Alert(ctx, kernel.Alert{Kind: "security", Title: "New sign-in", Body: "Firefox"})
	all, _ := uc.List(ctx, Filter{})
	if len(all) != 3 || all[0].Source != "security" || all[1].Message != "NAS is OK again" || all[2].Severity != domain.SeverityCritical {
		t.Fatalf("%+v", all)
	}
	if critical, _ := uc.List(ctx, Filter{Severity: domain.SeverityCritical}); len(critical) != 1 {
		t.Fatalf("filter: %+v", critical)
	}
	if d, _ := uc.RetentionDays(ctx); d != domain.DefaultRetentionDays {
		t.Fatal("default retention")
	}
	if err := uc.SetRetentionDays(ctx, 0); err == nil {
		t.Fatal("invalid retention accepted")
	}
	_ = uc.SetRetentionDays(ctx, 7)
	clk.t = clk.t.AddDate(0, 0, 8)
	_ = uc.Prune(ctx)
	if all, _ := uc.List(ctx, Filter{}); len(all) != 0 {
		t.Fatal("prune failed")
	}
}
