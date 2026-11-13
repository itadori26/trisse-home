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
	Provider    string
	Configured  bool
	URL         string
	Username    string
	HasPassword bool
}

// SettingsInput is the input of SaveSettings. An empty password keeps the
// stored one, so the UI never needs to know it.
type SettingsInput struct {
	Provider, URL, Username, Password string
}

// BlockedDomainView is a domain with its query count.
type BlockedDomainView struct {
	Domain string
	Count  int
}

// StatsView are the cached statistics of one provider.
type StatsView struct {
	Queries         int
	Blocked         int
	BlockedPercent  float64
	AvgProcessingMs float64
	// HasAvgProcessing is false for providers that do not report it.
	HasAvgProcessing  bool
	ProtectionEnabled bool
	TopBlocked        []BlockedDomainView
	UpdatedAt         time.Time
	Error             string
}

// DNSFilter is the input port of the module; provider is "adguard",
// "pihole" or empty (AdGuard Home).
type DNSFilter interface {
	Settings(ctx context.Context, provider string) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Stats(ctx context.Context, provider string) (StatsView, error)
	// Refresh updates every configured provider; it is run by the job.
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings per provider (output port).
// Load returns domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context, p domain.Provider) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// DNSFilterProvider queries one DNS filter (output port); AdGuard Home and
// Pi-hole are adapters implementing it.
type DNSFilterProvider interface {
	Stats(ctx context.Context, s domain.Settings) (domain.Stats, error)
}

type dnsFilter struct {
	repo      SettingsRepository
	providers map[domain.Provider]DNSFilterProvider
	clock     kernel.Clock
	mu        sync.RWMutex
	cached    map[domain.Provider]StatsView
}

// NewDNSFilter returns the use cases; providers maps each supported
// provider to its adapter.
func NewDNSFilter(repo SettingsRepository, providers map[domain.Provider]DNSFilterProvider, clock kernel.Clock) DNSFilter {
	return &dnsFilter{repo: repo, providers: providers, clock: clock, cached: map[domain.Provider]StatsView{}}
}

// Settings implements DNSFilter.
func (uc *dnsFilter) Settings(ctx context.Context, provider string) (SettingsView, error) {
	p, err := domain.ParseProvider(provider)
	if err != nil {
		return SettingsView{}, err
	}
	s, err := uc.repo.Load(ctx, p)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{Provider: string(p)}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return settingsView(s), nil
}

// SaveSettings implements DNSFilter. The connection is tested before saving
// so typos are reported immediately.
func (uc *dnsFilter) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	p, err := domain.ParseProvider(in.Provider)
	if err != nil {
		return SettingsView{}, err
	}
	password := in.Password
	if password == "" {
		if old, err := uc.repo.Load(ctx, p); err == nil {
			password = old.Password
		}
	}
	s, err := domain.NewSettings(p, in.URL, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	stats, err := uc.providers[p].Stats(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not connect to the DNS filter: {detail}", "detail", err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(p, toView(stats, uc.clock.Now()))
	return settingsView(s), nil
}

// Stats implements DNSFilter (cache-aside).
func (uc *dnsFilter) Stats(ctx context.Context, provider string) (StatsView, error) {
	p, err := domain.ParseProvider(provider)
	if err != nil {
		return StatsView{}, err
	}
	if v, ok := uc.cachedView(p); ok {
		return v, nil
	}
	s, err := uc.repo.Load(ctx, p)
	if err != nil {
		return StatsView{}, err
	}
	_ = uc.refresh(ctx, s) // a failed refresh is reported through StatsView.Error
	if v, ok := uc.cachedView(p); ok {
		return v, nil
	}
	return StatsView{}, kernel.NewPublicError(kernel.ErrUnavailable, "The DNS filter is not reachable.")
}

// Refresh implements DNSFilter.
func (uc *dnsFilter) Refresh(ctx context.Context) error {
	var errs []error
	for _, p := range domain.Providers {
		s, err := uc.repo.Load(ctx, p)
		if errors.Is(err, domain.ErrNotConfigured) {
			continue
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("load %s settings: %w", p, err))
			continue
		}
		errs = append(errs, uc.refresh(ctx, s))
	}
	return errors.Join(errs...)
}

func (uc *dnsFilter) refresh(ctx context.Context, s domain.Settings) error {
	stats, err := uc.providers[s.Provider].Stats(ctx, s)
	if err != nil {
		uc.mu.Lock()
		view := uc.cached[s.Provider]
		view.Error = "The DNS filter is not reachable."
		uc.cached[s.Provider] = view
		uc.mu.Unlock()
		return fmt.Errorf("query %s: %w", s.Provider, err)
	}
	uc.store(s.Provider, toView(stats, uc.clock.Now()))
	return nil
}

func (uc *dnsFilter) cachedView(p domain.Provider) (StatsView, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	v, ok := uc.cached[p]
	return v, ok
}

func (uc *dnsFilter) store(p domain.Provider, v StatsView) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached[p] = v
}

func settingsView(s domain.Settings) SettingsView {
	return SettingsView{Provider: string(s.Provider), Configured: true, URL: s.URL, Username: s.Username, HasPassword: s.Password != ""}
}

func toView(s domain.Stats, at time.Time) StatsView {
	v := StatsView{
		Queries: s.Queries, Blocked: s.Blocked, BlockedPercent: s.BlockedPercent(),
		AvgProcessingMs: float64(s.AvgProcessingTime.Microseconds()) / 1000, HasAvgProcessing: s.HasAvgProcessingTime,
		ProtectionEnabled: s.ProtectionEnabled, UpdatedAt: at, TopBlocked: make([]BlockedDomainView, 0, len(s.TopBlocked)),
	}
	for _, d := range s.TopBlocked {
		v.TopBlocked = append(v.TopBlocked, BlockedDomainView(d))
	}
	return v
}
