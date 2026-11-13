package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/shared/kernel"
)

// MaxDeals bounds the wishlist deals shown.
const MaxDeals = 15

// GameDeals is the input port of the module (#62).
type GameDeals interface {
	Settings(ctx context.Context) (domain.Settings, error)
	SaveSettings(ctx context.Context, in domain.Settings) (domain.Settings, error)
	Overview(ctx context.Context) (domain.Overview, error)
	// Refresh reloads both sources (job, daily).
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// WishlistProvider reads the discounted games of a public wishlist.
type WishlistProvider interface {
	Wishlist(ctx context.Context, steamID, country string) ([]domain.Deal, error)
}

// FreeGamesProvider reads the free game promotions.
type FreeGamesProvider interface {
	FreeGames(ctx context.Context, country string) ([]domain.FreeGame, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo     SettingsRepository
	Wishlist WishlistProvider
	Free     FreeGamesProvider
	Clock    kernel.Clock
}

type gameDeals struct {
	d      Deps
	mu     sync.RWMutex
	deals  []domain.Deal
	free   []domain.FreeGame
	errors []string
	loaded bool
}

// NewGameDeals returns the use cases.
func NewGameDeals(d Deps) GameDeals { return &gameDeals{d: d} }

// Settings implements GameDeals.
func (uc *gameDeals) Settings(ctx context.Context) (domain.Settings, error) {
	return uc.d.Repo.Load(ctx)
}

// SaveSettings implements GameDeals; the sources are loaded right away so
// a private wishlist is reported immediately.
func (uc *gameDeals) SaveSettings(ctx context.Context, in domain.Settings) (domain.Settings, error) {
	s, err := domain.NewSettings(in.SteamID, in.Country, in.MinDiscount, in.Epic)
	if err != nil {
		return domain.Settings{}, err
	}
	if s.SteamID != "" {
		if _, err := uc.d.Wishlist.Wishlist(ctx, s.SteamID, s.Country); err != nil {
			return domain.Settings{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not load the Steam wishlist (is it public?): {detail}", "detail", err.Error())
		}
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return domain.Settings{}, fmt.Errorf("save settings: %w", err)
	}
	_ = uc.Refresh(ctx) // errors are shown in the overview
	return s, nil
}

// Overview implements GameDeals (cache-aside).
func (uc *gameDeals) Overview(ctx context.Context) (domain.Overview, error) {
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return domain.Overview{}, err
	}
	uc.mu.RLock()
	loaded := uc.loaded
	uc.mu.RUnlock()
	if !loaded {
		_ = uc.Refresh(ctx)
	}
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	now := uc.d.Clock.Now()
	o := domain.Overview{Errors: slices.Clone(uc.errors), Deals: domain.FilterDeals(uc.deals, s.MinDiscount)}
	o.Deals = o.Deals[:min(len(o.Deals), MaxDeals)]
	for _, g := range uc.free {
		switch {
		case g.Current(now):
			o.FreeNow = append(o.FreeNow, g)
		case g.Start.After(now):
			o.FreeNext = append(o.FreeNext, g)
		}
	}
	return o, nil
}

// Refresh implements GameDeals; each source fails on its own.
func (uc *gameDeals) Refresh(ctx context.Context) error {
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	if !s.Enabled() {
		return nil
	}
	var deals []domain.Deal
	var free []domain.FreeGame
	var failed []string
	var errs []error
	if s.SteamID != "" {
		if deals, err = uc.d.Wishlist.Wishlist(ctx, s.SteamID, s.Country); err != nil {
			failed, errs = append(failed, "Steam"), append(errs, err)
		}
	}
	if s.Epic {
		if free, err = uc.d.Free.FreeGames(ctx, s.Country); err != nil {
			failed, errs = append(failed, "Epic Games Store"), append(errs, err)
		}
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if !slices.Contains(failed, "Steam") {
		uc.deals = deals
	}
	if !slices.Contains(failed, "Epic Games Store") {
		uc.free = free
	}
	uc.errors, uc.loaded = failed, true
	return errors.Join(errs...)
}

// RefreshInterval is how often the job runs: offers change at most daily.
const RefreshInterval = 12 * time.Hour
