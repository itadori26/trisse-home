package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/holidays/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error)   { return m.s, nil }
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = s; return nil }

type memCache struct {
	periods []domain.SchoolHoliday
	url     string
	at      time.Time
}

func (c *memCache) Load(context.Context) ([]domain.SchoolHoliday, string, time.Time, error) {
	return c.periods, c.url, c.at, nil
}
func (c *memCache) Store(_ context.Context, url string, p []domain.SchoolHoliday, at time.Time) error {
	c.periods, c.url, c.at = p, url, at
	return nil
}
func (c *memCache) Clear(context.Context) error { *c = memCache{}; return nil }

type fakeSource struct {
	calls int
	err   error
}

func (f *fakeSource) Fetch(context.Context, string, time.Time, time.Time) ([]domain.SchoolHoliday, error) {
	f.calls++
	d := func(m time.Month, day int) time.Time { return time.Date(2026, m, day, 0, 0, 0, 0, time.UTC) }
	return []domain.SchoolHoliday{{Name: "Weihnachtsferien", Start: d(12, 23), End: d(12, 31)}}, f.err
}

type clk struct{ now time.Time }

func (c *clk) Now() time.Time { return c.now }

func TestHolidays(t *testing.T) {
	ctx := context.Background()
	repo, cache, src, c := &memRepo{}, &memCache{}, &fakeSource{}, &clk{now: time.Date(2026, 12, 20, 10, 0, 0, 0, time.UTC)}
	uc := New(repo, src, cache, c)

	// Without a link only public holidays, and nothing is requested.
	o, err := uc.Upcoming(ctx, "NI", 3)
	if err != nil || o.School || len(o.Entries) != 3 || o.Entries[0].Key != "christmas_day" || src.calls != 0 {
		t.Fatalf("%+v %v calls=%d", o, err, src.calls)
	}
	if err := uc.Refresh(ctx); err != nil || src.calls != 0 {
		t.Fatalf("refresh without link: %v %d", err, src.calls)
	}

	// A broken link is reported and not saved.
	src.err = errors.New("404")
	if _, err := uc.SaveSettings(ctx, "https://example.org/ferien.ics"); !errors.Is(err, kernel.ErrInvalid) || repo.s.SchoolURL != "" {
		t.Fatalf("broken link: %v %+v", err, repo.s)
	}
	src.err = nil
	if _, err := uc.SaveSettings(ctx, "https://example.org/ferien.ics"); err != nil {
		t.Fatal(err)
	}
	o, _ = uc.Upcoming(ctx, "NI", 3)
	if !o.School || o.Entries[0].Name != "Weihnachtsferien" || o.SchoolUpdatedAt.IsZero() {
		t.Fatalf("%+v", o)
	}

	// The cache is used for a week, then the job loads again.
	calls := src.calls
	c.now = c.now.Add(6 * 24 * time.Hour)
	_ = uc.Refresh(ctx)
	if src.calls != calls {
		t.Fatal("reloaded too early")
	}
	c.now = c.now.Add(2 * 24 * time.Hour)
	src.err = errors.New("down")
	if err := uc.Refresh(ctx); err == nil {
		t.Fatal("expected an error")
	}
	o, _ = uc.Upcoming(ctx, "NI", 10)
	if o.SchoolError == "" || len(cache.periods) != 1 {
		t.Fatalf("last known dates must stay: %+v", o)
	}

	// Removing the link clears the cache.
	if _, err := uc.SaveSettings(ctx, " "); err != nil || cache.url != "" || repo.s.SchoolURL != "" {
		t.Fatalf("remove: %v %+v", err, cache)
	}
}

func TestEmptyCalendarIsRejected(t *testing.T) {
	src := &emptySource{}
	uc := New(&memRepo{}, src, &memCache{}, &clk{now: time.Now()})
	if _, err := uc.SaveSettings(context.Background(), "https://example.org/x.ics"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("%v", err)
	}
}

type emptySource struct{}

func (emptySource) Fetch(context.Context, string, time.Time, time.Time) ([]domain.SchoolHoliday, error) {
	return nil, nil
}
