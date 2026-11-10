package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeSource struct{ fail map[string]bool }

var now = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func (f *fakeSource) Events(_ context.Context, s domain.Settings, u string, _, _ time.Time) ([]domain.Event, error) {
	if s.Password != "pw" {
		return nil, errors.New("401 Unauthorized")
	}
	if f.fail[u] {
		return nil, errors.New("404 Not Found")
	}
	name := domain.CalendarName(u)
	return []domain.Event{
		{Summary: name + " past", Start: now.Add(-3 * time.Hour), End: now.Add(-2 * time.Hour), Calendar: name},
		{Summary: name + " soon", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour), Calendar: name},
		{Summary: name + " later", Start: now.Add(10 * 24 * time.Hour), End: now.Add(10*24*time.Hour + time.Hour), Calendar: name},
	}, nil
}

type clk struct{}

func (clk) Now() time.Time { return now }

func TestCalendar(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{fail: map[string]bool{}}
	uc := NewCalendar(repo, src, clk{})
	if _, err := uc.Upcoming(ctx, 7); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	urls := []string{"https://c.example/cal/home/", "https://c.example/cal/work/"}
	if _, err := uc.SaveSettings(ctx, SettingsInput{CalendarURLs: urls, Username: "me", Password: "bad"}); err == nil {
		t.Fatal("expected connection error")
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{CalendarURLs: urls, Username: "me", Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	week, _ := uc.Upcoming(ctx, 7)
	if len(week.Events) != 2 || !strings.HasSuffix(week.Events[0].Summary, "soon") {
		t.Fatalf("7-day agenda wrong: %+v", week.Events)
	}
	month, _ := uc.Upcoming(ctx, 31)
	if len(month.Events) != 4 {
		t.Fatalf("31-day agenda wrong: %d events", len(month.Events))
	}
	src.fail["https://c.example/cal/work/"] = true
	if err := uc.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	partial, _ := uc.Upcoming(ctx, 31)
	if len(partial.Events) != 2 || !strings.Contains(partial.Error, "work") {
		t.Fatalf("partial failure not reported: %+v", partial)
	}
}
