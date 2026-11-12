package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/adguard/domain"
	"trisse-home/internal/shared/kernel"
)

// SettingsView never contains the password, only whether one is stored.
type SettingsView struct {
	Configured  bool
	URL         string
	Username    string
	HasPassword bool
}

// SettingsInput is the input of SaveSettings. An empty password keeps the
// stored one, so the UI never needs to know it.
type SettingsInput struct {
	URL, Username, Password string
}

// BlockedDomainView is a domain with its query count.
type BlockedDomainView struct {
	Domain string
	Count  int
}

// StatsView are the cached statistics.
type StatsView struct {
	Queries           int
	Blocked           int
	BlockedPercent    float64
	AvgProcessingMs   float64
	ProtectionEnabled bool
	TopBlocked        []BlockedDomainView
	UpdatedAt         time.Time
	Error             string
}

// AdGuard is the input port of the module.
type AdGuard interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Stats(ctx context.Context) (StatsView, error)
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port). Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// StatsSource queries AdGuard Home (output port).
type StatsSource interface {
	Stats(ctx context.Context, s domain.Settings) (domain.Stats, error)
}

type adguard struct {
	repo   SettingsRepository
	source StatsSource
	clock  kernel.Clock
	mu     sync.RWMutex
	cached *StatsView
}

// NewAdGuard returns the AdGuard use cases.
func NewAdGuard(repo SettingsRepository, source StatsSource, clock kernel.Clock) AdGuard {
	return &adguard{repo: repo, source: source, clock: clock}
}

// Settings implements AdGuard.
func (uc *adguard) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: s.Password != ""}, nil
}

// SaveSettings implements AdGuard. The connection is tested before saving so
// typos are reported immediately.
func (uc *adguard) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	password := in.Password
	if password == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			password = old.Password
		}
	}
	s, err := domain.NewSettings(in.URL, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	stats, err := uc.source.Stats(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not connect to AdGuard Home: {detail}", "detail", err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(toView(stats, uc.clock.Now()))
	return SettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: true}, nil
}

// Stats implements AdGuard (cache-aside).
func (uc *adguard) Stats(ctx context.Context) (StatsView, error) {
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()
	if cached != nil {
		return *cached, nil
	}
	if _, err := uc.repo.Load(ctx); err != nil {
		return StatsView{}, err
	}
	_ = uc.Refresh(ctx) // a failed refresh is reported through StatsView.Error
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	if uc.cached == nil {
		return StatsView{}, kernel.NewPublicError(kernel.ErrUnavailable, "AdGuard Home is not reachable.")
	}
	return *uc.cached, nil
}

// Refresh implements AdGuard; it is run by the background job.
func (uc *adguard) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	stats, err := uc.source.Stats(ctx, s)
	if err != nil {
		uc.mu.Lock()
		view := StatsView{}
		if uc.cached != nil {
			view = *uc.cached
		}
		view.Error = "AdGuard Home is not reachable."
		uc.cached = &view
		uc.mu.Unlock()
		return fmt.Errorf("query adguard: %w", err)
	}
	uc.store(toView(stats, uc.clock.Now()))
	return nil
}

func (uc *adguard) store(v StatsView) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = &v
}

func toView(s domain.Stats, at time.Time) StatsView {
	v := StatsView{
		Queries: s.Queries, Blocked: s.Blocked, BlockedPercent: s.BlockedPercent(),
		AvgProcessingMs: float64(s.AvgProcessingTime.Microseconds()) / 1000, ProtectionEnabled: s.ProtectionEnabled,
		UpdatedAt: at, TopBlocked: make([]BlockedDomainView, 0, len(s.TopBlocked)),
	}
	for _, d := range s.TopBlocked {
		v.TopBlocked = append(v.TopBlocked, BlockedDomainView(d))
	}
	return v
}
