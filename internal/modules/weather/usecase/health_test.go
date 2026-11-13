package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeHealth struct {
	pollenCalls, uvCalls int
	err                  error
}

func (f *fakeHealth) Pollen(context.Context) ([]domain.PollenForecast, error) {
	f.pollenCalls++
	return []domain.PollenForecast{{Region: domain.PollenRegion{ID: 32, Name: "Östl. Niedersachsen"},
		Days: []domain.PollenDay{{Type: domain.Hazel, Today: 2}, {Type: domain.Birch, Today: 0}}}}, f.err
}

func (f *fakeHealth) UV(context.Context) ([]domain.UVForecast, error) {
	f.uvCalls++
	return []domain.UVForecast{{City: "Hannover", Today: 4}, {City: "Bremen", Today: 3}}, f.err
}

type stepClock struct{ now time.Time }

func (c *stepClock) Now() time.Time { return c.now }

func TestHealth(t *testing.T) {
	ctx := context.Background()
	p, c := &fakeHealth{}, &stepClock{now: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)}
	uc := NewHealth(p, c)

	// Nothing asked for, nothing requested.
	if r, err := uc.Report(ctx, 0, "", nil); err != nil || r.Pollen != nil || r.UV != nil || p.pollenCalls+p.uvCalls != 0 {
		t.Fatalf("%+v %v", r, err)
	}
	r, err := uc.Report(ctx, 32, "Hannover", []domain.PollenType{domain.Hazel})
	if err != nil || r.Pollen == nil || len(r.Pollen.Days) != 1 || r.Pollen.Days[0].Type != domain.Hazel || r.UV == nil || r.UV.Today != 4 {
		t.Fatalf("%+v %v", r, err)
	}
	// Cached for an hour, shared by all widgets.
	_, _ = uc.Report(ctx, 32, "Bremen", nil)
	if p.pollenCalls != 1 || p.uvCalls != 1 {
		t.Fatalf("calls %d %d", p.pollenCalls, p.uvCalls)
	}
	// Unknown region or city: that part is empty.
	if r, _ := uc.Report(ctx, 99, "Atlantis", nil); r.Pollen != nil || r.UV != nil {
		t.Fatalf("%+v", r)
	}
	o, err := uc.Options(ctx)
	if err != nil || len(o.Regions) != 1 || o.Cities[0] != "Bremen" {
		t.Fatalf("%+v %v", o, err)
	}
	// After an hour it is loaded again; a failure keeps the last forecast.
	c.now = c.now.Add(61 * time.Minute)
	p.err = errors.New("down")
	if r, err := uc.Report(ctx, 32, "", nil); err != nil || r.Pollen == nil || p.pollenCalls != 2 {
		t.Fatalf("stale: %+v %v", r, err)
	}
	// Never loaded and failing: unavailable.
	fresh := NewHealth(&fakeHealth{err: errors.New("down")}, c)
	if _, err := fresh.Report(ctx, 32, "", nil); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("%v", err)
	}
}
