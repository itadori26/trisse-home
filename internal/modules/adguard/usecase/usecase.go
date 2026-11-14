package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	// CanPause is true when DNS_PAUSE_PROVIDERS allows pausing (#177).
	CanPause bool
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
	// PausedUntil is set while a timed pause runs (#177).
	PausedUntil *time.Time
	TopBlocked  []BlockedDomainView
	UpdatedAt   time.Time
	Error       string
}

// DNSFilter is the input port of the module; provider is "adguard",
// "pihole" or empty (AdGuard Home).
type DNSFilter interface {
	Settings(ctx context.Context, provider string) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Stats(ctx context.Context, provider string) (StatsView, error)
	// Refresh updates every configured provider; it is run by the job.
	Refresh(ctx context.Context) error
	// Pause switches blocking off for a few minutes, if the server allows it
	// for the provider (#177); it returns the new statistics.
	Pause(ctx context.Context, provider string, minutes int) (StatsView, error)
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
	// Pause switches blocking off for d; the filter turns itself back on.
	Pause(ctx context.Context, s domain.Settings, d time.Duration) error
}

// Deps are the dependencies of the use cases.
type Deps struct {
	Repo SettingsRepository
	// Providers maps each supported provider to its adapter.
	Providers map[domain.Provider]DNSFilterProvider
	// Pausable lists the providers that may be paused (DNS_PAUSE_PROVIDERS).
	Pausable []domain.Provider
	Audit    kernel.AuditSink
	Clock    kernel.Clock
}

type dnsFilter struct {
	repo      SettingsRepository
	providers map[domain.Provider]DNSFilterProvider
	pausable  []domain.Provider
	audit     kernel.AuditSink
	clock     kernel.Clock
	mu        sync.RWMutex
	cached    map[domain.Provider]StatsView
}

// NewDNSFilter returns the use cases.
func NewDNSFilter(d Deps) DNSFilter {
	if d.Audit == nil {
		d.Audit = kernel.DiscardAudit{}
	}
	return &dnsFilter{repo: d.Repo, providers: d.Providers, pausable: d.Pausable, audit: d.Audit, clock: d.Clock, cached: map[domain.Provider]StatsView{}}
}

// Settings implements DNSFilter.
func (uc *dnsFilter) Settings(ctx context.Context, provider string) (SettingsView, error) {
	p, err := domain.ParseProvider(provider)
	if err != nil {
		return SettingsView{}, err
	}
	s, err := uc.repo.Load(ctx, p)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{Provider: string(p), CanPause: uc.canPause(p)}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return uc.settingsView(s), nil
}

func (uc *dnsFilter) canPause(p domain.Provider) bool { return slices.Contains(uc.pausable, p) }

// Pause implements DNSFilter.
func (uc *dnsFilter) Pause(ctx context.Context, provider string, minutes int) (StatsView, error) {
	p, err := domain.ParseProvider(provider)
	if err != nil {
		return StatsView{}, err
	}
	if !uc.canPause(p) {
		return StatsView{}, domain.ErrPauseNotAllowed
	}
	d, err := domain.ParsePause(minutes)
	if err != nil {
		return StatsView{}, err
	}
	s, err := uc.repo.Load(ctx, p)
	if err != nil {
		return StatsView{}, err
	}
	if err := uc.providers[p].Pause(ctx, s, d); err != nil {
		return StatsView{}, kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	uc.audit.Audit(ctx, kernel.AuditEvent{Type: "dns_filter_paused", Detail: fmt.Sprintf("%s for %d min", p, minutes)})
	if err := uc.refresh(ctx, s); err != nil {
		// Paused, but the statistics could not be read back: report the pause.
		until := uc.clock.Now().Add(d)
		v, _ := uc.cachedView(p)
		v.ProtectionEnabled, v.PausedUntil = false, &until
		uc.store(p, v)
	}
	v, _ := uc.cachedView(p)
	return v, nil
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
			// The stored password only goes to the same server and user (#191).
			if kernel.TargetChanged([]string{old.URL, old.Username}, []string{in.URL, in.Username}) {
				return SettingsView{}, kernel.SecretRequired("password")
			}
			password = old.Password
		}
	}
	s, err := domain.NewSettings(p, in.URL, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	stats, err := uc.providers[p].Stats(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(p, toView(stats, uc.clock.Now()))
	return uc.settingsView(s), nil
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

func (uc *dnsFilter) settingsView(s domain.Settings) SettingsView {
	return SettingsView{Provider: string(s.Provider), Configured: true, URL: s.URL, Username: s.Username, HasPassword: s.Password != "",
		CanPause: uc.canPause(s.Provider)}
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
	if !s.ProtectionEnabled && s.PauseRemaining > 0 {
		until := at.Add(s.PauseRemaining)
		v.PausedUntil = &until
	}
	return v
}
