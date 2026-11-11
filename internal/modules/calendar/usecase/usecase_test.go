package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
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
	uc := NewCalendar(repo, src, &fakeWriter{}, &seqIDs{}, clk{})
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

type fakeWriter struct {
	created []domain.Event
	urls    []string
	err     error
}

func (f *fakeWriter) Create(_ context.Context, _ domain.Settings, u string, e domain.Event) error {
	f.created, f.urls = append(f.created, e), append(f.urls, u)
	return f.err
}

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return "id" + strings.Repeat("x", s.n) }

func TestRangeAndCreate(t *testing.T) {
	ctx := context.Background()
	repo, src, w := &memRepo{}, &fakeSource{fail: map[string]bool{}}, &fakeWriter{}
	uc := NewCalendar(repo, src, w, &seqIDs{}, clk{})
	urls := []string{"https://c.example/cal/home/"}
	if _, err := uc.SaveSettings(ctx, SettingsInput{CalendarURLs: urls, Username: "me", Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Range(ctx, now, now.AddDate(0, 0, 60)); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("too long range accepted: %v", err)
	}
	inCache, err := uc.Range(ctx, now, now.AddDate(0, 0, 7))
	if err != nil || len(inCache.Events) == 0 {
		t.Fatalf("cached range: %+v %v", inCache, err)
	}
	past, err := uc.Range(ctx, now.AddDate(0, -2, 0), now.AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("live range: %v", err)
	}
	_ = past
	if _, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: "https://evil.example/", Summary: "x", Start: now, End: now.Add(time.Hour)}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown calendar accepted: %v", err)
	}
	day := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	v, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: urls[0], Summary: "Holiday", AllDay: true, Start: day, End: day})
	if err != nil || !v.End.Equal(day.AddDate(0, 0, 1)) || !strings.HasSuffix(w.created[0].UID, "@trisse-home") || v.Calendar != "home" {
		t.Fatalf("all-day create: %+v %v %+v", v, err, w.created)
	}
	w.err = errors.New("403 Forbidden")
	if _, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: urls[0], Summary: "x", Start: now, End: now.Add(time.Hour)}); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("server error not reported: %v", err)
	}
}
