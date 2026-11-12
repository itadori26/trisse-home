package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/shared/kernel"
)

// Overview is returned by Upcoming.
type Overview struct {
	Types     []domain.Upcoming
	UpdatedAt time.Time
	Error     string
}

// Waste is the input port of the module.
type Waste interface {
	Settings(ctx context.Context) (domain.Settings, bool, error)
	SaveSettings(ctx context.Context, rawURL string, rules []domain.Rule) (domain.Settings, error)
	Upcoming(ctx context.Context) (Overview, error)
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// CollectionSource loads the collection dates of an ICS URL (output port).
type CollectionSource interface {
	Collections(ctx context.Context, url string, from, to time.Time) ([]domain.Collection, error)
}

// horizon is how far ahead dates are loaded.
const horizon = 366 * 24 * time.Hour

type waste struct {
	repo   SettingsRepository
	source CollectionSource
	clock  kernel.Clock

	mu        sync.RWMutex
	loaded    bool
	cs        []domain.Collection
	updatedAt time.Time
	lastError string
}

// NewWaste returns the waste use cases.
func NewWaste(repo SettingsRepository, source CollectionSource, clock kernel.Clock) Waste {
	return &waste{repo: repo, source: source, clock: clock}
}

// Settings implements Waste; configured is false without settings.
func (uc *waste) Settings(ctx context.Context) (domain.Settings, bool, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return domain.Settings{Rules: []domain.Rule{}}, false, nil
	}
	if err != nil {
		return domain.Settings{}, false, fmt.Errorf("load settings: %w", err)
	}
	return s, true, nil
}

// SaveSettings implements Waste; the subscription is loaded first so typos
// are reported immediately.
func (uc *waste) SaveSettings(ctx context.Context, rawURL string, rules []domain.Rule) (domain.Settings, error) {
	s, err := domain.NewSettings(rawURL, rules)
	if err != nil {
		return domain.Settings{}, err
	}
	now := uc.clock.Now()
	cs, err := uc.source.Collections(ctx, s.URL, now.Add(-24*time.Hour), now.Add(horizon))
	if err != nil {
		return domain.Settings{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not load the waste calendar: {detail}", "detail", err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return domain.Settings{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(cs, "", now)
	return s, nil
}

// Upcoming implements Waste.
func (uc *waste) Upcoming(ctx context.Context) (Overview, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return Overview{}, err
	}
	uc.mu.RLock()
	loaded := uc.loaded
	uc.mu.RUnlock()
	if !loaded {
		_ = uc.Refresh(ctx) // failures are reported through Overview.Error
	}
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return Overview{
		Types: domain.NextCollections(slices.Clone(uc.cs), s.Rules, uc.clock.Now()), UpdatedAt: uc.updatedAt, Error: uc.lastError,
	}, nil
}

// Refresh implements Waste; it is run by the background job.
func (uc *waste) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	now := uc.clock.Now()
	cs, err := uc.source.Collections(ctx, s.URL, now.Add(-24*time.Hour), now.Add(horizon))
	if err != nil {
		uc.mu.Lock()
		uc.loaded, uc.lastError = true, "The waste calendar could not be loaded; showing the last known dates."
		uc.mu.Unlock()
		return fmt.Errorf("load waste calendar: %w", err)
	}
	uc.store(cs, "", now)
	return nil
}

func (uc *waste) store(cs []domain.Collection, errMsg string, at time.Time) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.loaded, uc.cs, uc.lastError, uc.updatedAt = true, cs, errMsg, at
}
