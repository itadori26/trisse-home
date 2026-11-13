package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/fuel/domain"
	"trisse-home/internal/shared/kernel"
)

const (
	key  = "00000000-0000-0000-0000-000000000002"
	near = "11111111-1111-1111-1111-111111111111"
	far  = "22222222-2222-2222-2222-222222222222"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeProvider struct {
	around, prices int
	err            error
	asked          []string
}

func (f *fakeProvider) Around(_ context.Context, s domain.Settings) ([]domain.Station, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.around++
	return []domain.Station{{ID: near, Name: "Near", Price: 1.79, Open: true}, {ID: "x", Name: "Cheap", Price: 1.69, Open: true}}, nil
}

func (f *fakeProvider) Prices(_ context.Context, _ string, ids []string) (map[string]domain.Prices, error) {
	f.prices++
	f.asked = ids
	return map[string]domain.Prices{far: {Open: true, ByFuel: map[domain.FuelType]float64{domain.FuelE10: 1.65}}}, nil
}

type movingClock struct{ now time.Time }

func (c *movingClock) Now() time.Time { return c.now }

func TestFuel(t *testing.T) {
	ctx := context.Background()
	repo, p := &memRepo{}, &fakeProvider{}
	c := &movingClock{now: time.Unix(0, 0)}
	uc := NewFuel(repo, p, c)

	if v, _ := uc.Settings(ctx); v.Configured || v.Fuel != "e10" || v.RadiusKm != 5 {
		t.Fatalf("defaults: %+v", v)
	}
	if _, err := uc.Overview(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}
	p.err = errors.New("apikey invalid")
	if _, err := uc.SaveSettings(ctx, SettingsInput{APIKey: key, Lat: 52.5, Lon: 13.4, RadiusKm: 5, Fuel: "e10"}); !errors.Is(err, kernel.ErrInvalid) || repo.s != nil {
		t.Fatalf("a failing test must not save: %v", err)
	}
	p.err = nil
	if _, err := uc.SaveSettings(ctx, SettingsInput{APIKey: key, Lat: 52.5, Lon: 13.4, RadiusKm: 5, Fuel: "e10"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.SetFavourites(ctx, []domain.Favourite{{ID: near, Name: "Near"}, {ID: far, Name: "Far away"}}); err != nil {
		t.Fatal(err)
	}
	if v, err := uc.SaveSettings(ctx, SettingsInput{Lat: 52.5, Lon: 13.4, RadiusKm: 10, Fuel: "e10"}); err != nil || len(v.Favourites) != 2 || repo.s.APIKey != key {
		t.Fatalf("an empty key and the favourites are kept: %+v %v", v, err)
	}

	o, err := uc.Overview(ctx)
	if err != nil || o.Cheapest[0].Name != "Cheap" || len(o.Favourites) != 2 {
		t.Fatalf("overview: %+v %v", o, err)
	}
	if o.Favourites[1].Name != "Far away" || o.Favourites[1].Price != 1.65 || len(p.asked) != 1 || p.asked[0] != far {
		t.Fatalf("only favourites outside the radius are asked for: %+v %v", o.Favourites, p.asked)
	}
	calls := p.around
	_, _ = uc.Overview(ctx)
	c.now = c.now.Add(4 * time.Minute)
	_, _ = uc.Overview(ctx)
	if p.around != calls {
		t.Fatal("at most one request every five minutes")
	}
	c.now = c.now.Add(time.Minute)
	_, _ = uc.Overview(ctx)
	if p.around != calls+1 {
		t.Fatal("refreshed after five minutes")
	}
	p.err = errors.New("down")
	c.now = c.now.Add(cacheTTL)
	if _, err := uc.Overview(ctx); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("unreachable: %v", err)
	}
}
