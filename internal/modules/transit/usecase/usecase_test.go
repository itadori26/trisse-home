package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/transit/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeProvider struct {
	calls int
	err   error
}

func (f *fakeProvider) SearchStops(context.Context, string) ([]domain.Stop, error) {
	return []domain.Stop{{ID: "s1", Name: "Hbf"}}, f.err
}

func (f *fakeProvider) Departures(context.Context, string, int) ([]domain.Departure, error) {
	f.calls++
	base := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	return []domain.Departure{{Line: "1", Planned: base.Add(5 * time.Minute)}, {Line: "2", Planned: base.Add(9 * time.Minute)}}, f.err
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

func TestTransit(t *testing.T) {
	ctx := context.Background()
	p, c := &fakeProvider{}, &clk{t: time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)}
	uc := NewTransit(p, c)
	if s, _ := uc.SearchStops(ctx, "H"); len(s) != 0 {
		t.Fatal("one-letter search must not query the provider")
	}
	if _, err := uc.Departures(ctx, "", 5); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("empty stop accepted")
	}
	d, err := uc.Departures(ctx, "s1", 1)
	if err != nil || len(d) != 1 || d[0].Line != "1" {
		t.Fatalf("%+v %v", d, err)
	}
	_, _ = uc.Departures(ctx, "s1", 5)
	if p.calls != 1 {
		t.Fatalf("cache not used: %d", p.calls)
	}
	c.t = c.t.Add(2 * time.Minute)
	p.err = errors.New("429")
	if d, err := uc.Departures(ctx, "s1", 5); err != nil || len(d) != 2 {
		t.Fatalf("stale data must beat an error: %v %v", d, err)
	}
	if _, err := uc.Departures(ctx, "other", 5); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("error without cache: %v", err)
	}
}
