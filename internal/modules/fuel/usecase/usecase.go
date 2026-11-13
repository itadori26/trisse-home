package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/fuel/domain"
	"trisse-home/internal/shared/kernel"
)

// cacheTTL follows the provider's terms: at most one request every five
// minutes, and only while a widget is open.
const cacheTTL = 5 * time.Minute

// SettingsView never contains the API key.
type SettingsView struct {
	Configured bool
	HasAPIKey  bool
	Lat, Lon   float64
	Place      string
	RadiusKm   float64
	Fuel       string
	Favourites []domain.Favourite
}

// SettingsInput is the input of SaveSettings; an empty key keeps the stored one.
type SettingsInput struct {
	APIKey   string
	Lat, Lon float64
	Place    string
	RadiusKm float64
	Fuel     string
}

// Fuel is the input port of the module (#53).
type Fuel interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	SetFavourites(ctx context.Context, list []domain.Favourite) (SettingsView, error)
	Overview(ctx context.Context) (domain.Overview, error)
}

// SettingsRepository persists the settings (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// FuelPriceProvider reads prices (output port); Tankerkönig is the adapter.
type FuelPriceProvider interface {
	// Around lists the stations within the radius with the price of s.Fuel.
	Around(ctx context.Context, s domain.Settings) ([]domain.Station, error)
	// Prices returns the prices of up to ten stations by ID.
	Prices(ctx context.Context, apiKey string, ids []string) (map[string]domain.Prices, error)
}

type fuel struct {
	repo     SettingsRepository
	provider FuelPriceProvider
	clock    kernel.Clock
	mu       sync.Mutex
	cached   *domain.Overview
	fetched  time.Time
}

// NewFuel returns the use cases.
func NewFuel(repo SettingsRepository, provider FuelPriceProvider, clock kernel.Clock) Fuel {
	return &fuel{repo: repo, provider: provider, clock: clock}
}

// Settings implements Fuel.
func (uc *fuel) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{RadiusKm: 5, Fuel: string(domain.FuelE10), Favourites: []domain.Favourite{}}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return view(s), nil
}

// SaveSettings implements Fuel; the key and location are tested first.
func (uc *fuel) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	old, oldErr := uc.repo.Load(ctx)
	key := in.APIKey
	if key == "" && oldErr == nil {
		key = old.APIKey
	}
	s, err := domain.NewSettings(key, in.Lat, in.Lon, in.Place, in.RadiusKm, in.Fuel)
	if err != nil {
		return SettingsView{}, err
	}
	if oldErr == nil {
		s.Favourites = old.Favourites
	}
	if _, err := uc.provider.Around(ctx, s); err != nil {
		return SettingsView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not load fuel prices: {detail}", "detail", err.Error())
	}
	return uc.save(ctx, s)
}

// SetFavourites implements Fuel.
func (uc *fuel) SetFavourites(ctx context.Context, list []domain.Favourite) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	if err := s.SetFavourites(list); err != nil {
		return SettingsView{}, err
	}
	return uc.save(ctx, s)
}

func (uc *fuel) save(ctx context.Context, s domain.Settings) (SettingsView, error) {
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	uc.cached = nil
	uc.mu.Unlock()
	return view(s), nil
}

// Overview implements Fuel. A failed favourites request keeps the
// cheapest list; the provider's errors never hide everything.
func (uc *fuel) Overview(ctx context.Context) (domain.Overview, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Overview{}, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.clock.Now()
	if uc.cached != nil && now.Sub(uc.fetched) < cacheTTL {
		return *uc.cached, nil
	}
	around, err := uc.provider.Around(ctx, s)
	if err != nil {
		return domain.Overview{}, kernel.NewPublicError(kernel.ErrUnavailable, "The fuel price service is not reachable.")
	}
	domain.SortCheapest(around)
	o := domain.Overview{Fuel: s.Fuel, Place: s.Place, Cheapest: around[:min(len(around), domain.MaxCheapest)]}
	o.Favourites = uc.favourites(ctx, s, around)
	uc.cached, uc.fetched = &o, now
	return o, nil
}

// favourites takes stations from the radius search and asks the provider
// only for those outside of it.
func (uc *fuel) favourites(ctx context.Context, s domain.Settings, around []domain.Station) []domain.Station {
	byID := make(map[string]domain.Station, len(around))
	for _, st := range around {
		byID[st.ID] = st
	}
	var missing []string
	for _, f := range s.Favourites {
		if _, ok := byID[f.ID]; !ok {
			missing = append(missing, f.ID)
		}
	}
	prices := map[string]domain.Prices{}
	if len(missing) > 0 {
		prices, _ = uc.provider.Prices(ctx, s.APIKey, missing) // best effort; the station then shows no price
	}
	out := make([]domain.Station, 0, len(s.Favourites))
	for _, f := range s.Favourites {
		if st, ok := byID[f.ID]; ok {
			out = append(out, st)
			continue
		}
		p := prices[f.ID]
		out = append(out, domain.Station{ID: f.ID, Name: f.Name, Open: p.Open, Price: p.ByFuel[s.Fuel]})
	}
	return out
}

func view(s domain.Settings) SettingsView {
	favs := s.Favourites
	if favs == nil {
		favs = []domain.Favourite{}
	}
	return SettingsView{Configured: true, HasAPIKey: s.APIKey != "", Lat: s.Lat, Lon: s.Lon, Place: s.Place, RadiusKm: s.RadiusKm,
		Fuel: string(s.Fuel), Favourites: favs}
}
