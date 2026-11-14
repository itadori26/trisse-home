package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/adguard/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	s map[domain.Provider]domain.Settings
}

func (m *memRepo) Load(_ context.Context, p domain.Provider) (domain.Settings, error) {
	s, ok := m.s[p]
	if !ok {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s[s.Provider] = s; return nil }

type fakeSource struct {
	err    error
	stats  domain.Stats
	paused time.Duration
}

func (f *fakeSource) Pause(_ context.Context, _ domain.Settings, d time.Duration) error {
	if f.err != nil {
		return f.err
	}
	f.paused = d
	f.stats.ProtectionEnabled, f.stats.PauseRemaining = false, d
	return nil
}

type audit struct{ events []kernel.AuditEvent }

func (a *audit) Audit(_ context.Context, e kernel.AuditEvent) { a.events = append(a.events, e) }

func (f *fakeSource) Stats(_ context.Context, s domain.Settings) (domain.Stats, error) {
	if f.err != nil {
		return domain.Stats{}, f.err
	}
	if s.Password != "secret" {
		return domain.Stats{}, errors.New("401 Unauthorized")
	}
	return f.stats, nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func setup() (DNSFilter, *memRepo, *fakeSource, *fakeSource) {
	repo := &memRepo{s: map[domain.Provider]domain.Settings{}}
	ag := &fakeSource{stats: domain.Stats{Queries: 100, Blocked: 20, AvgProcessingTime: 1500 * time.Microsecond, HasAvgProcessingTime: true}}
	pi := &fakeSource{stats: domain.Stats{Queries: 400, Blocked: 100, ProtectionEnabled: true}}
	uc := NewDNSFilter(Deps{Repo: repo, Providers: map[domain.Provider]DNSFilterProvider{domain.ProviderAdGuard: ag, domain.ProviderPiHole: pi}, Clock: clk{}})
	return uc, repo, ag, pi
}

func TestAdGuard(t *testing.T) {
	ctx := context.Background()
	uc, repo, src, _ := setup()

	if s, _ := uc.Settings(ctx, ""); s.Configured || s.Provider != "adguard" {
		t.Fatalf("expected unconfigured adguard: %+v", s)
	}
	if _, err := uc.Stats(ctx, ""); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ag:3000", Username: "admin", Password: "wrong"}); err == nil {
		t.Fatal("expected connection test to fail")
	}
	if len(repo.s) != 0 {
		t.Fatal("invalid settings must not be stored")
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ag:3000", Username: "admin", Password: "secret"}); err != nil {
		t.Fatal(err)
	}
	// An empty password keeps the stored one for the same server and user …
	view, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ag:3000/", Username: "admin"})
	if err != nil || !view.HasPassword || repo.s[domain.ProviderAdGuard].Password != "secret" {
		t.Fatalf("password not kept: %+v %v", view, err)
	}
	// … but is never sent to a changed server or user (#191).
	for _, in := range []SettingsInput{{URL: "http://ag:3000", Username: "admin2"}, {URL: "http://other:3000", Username: "admin"}} {
		var v *kernel.ValidationError
		if _, err := uc.SaveSettings(ctx, in); !errors.As(err, &v) || v.Field != "password" {
			t.Fatalf("%+v: %v", in, err)
		}
	}
	stats, _ := uc.Stats(ctx, "adguard")
	if stats.BlockedPercent != 20 || stats.AvgProcessingMs != 1.5 || !stats.HasAvgProcessing {
		t.Fatalf("stats %+v", stats)
	}
	src.err = errors.New("timeout")
	_ = uc.Refresh(ctx)
	if stats, _ := uc.Stats(ctx, ""); stats.Error == "" || stats.Queries != 100 {
		t.Fatalf("last known stats not kept: %+v", stats)
	}
}

func TestPiHoleBesideAdGuard(t *testing.T) {
	ctx := context.Background()
	uc, _, ag, pi := setup()
	if _, err := uc.SaveSettings(ctx, SettingsInput{Provider: "pihole", URL: "http://pi.hole", Password: "secret"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Stats(ctx, "adguard"); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("AdGuard stays unconfigured: %v", err)
	}
	stats, err := uc.Stats(ctx, "pihole")
	if err != nil || stats.BlockedPercent != 25 || stats.HasAvgProcessing || !stats.ProtectionEnabled {
		t.Fatalf("pi-hole stats: %+v %v", stats, err)
	}
	if s, _ := uc.Settings(ctx, "pihole"); !s.Configured || s.Provider != "pihole" || !s.HasPassword {
		t.Fatalf("pi-hole settings: %+v", s)
	}
	pi.stats.Queries = 800
	ag.err = errors.New("must not be asked")
	if err := uc.Refresh(ctx); err != nil {
		t.Fatalf("only configured providers are refreshed: %v", err)
	}
	if stats, _ := uc.Stats(ctx, "pihole"); stats.Queries != 800 {
		t.Fatalf("refresh: %+v", stats)
	}
	pi.err = errors.New("down")
	if err := uc.Refresh(ctx); err == nil {
		t.Fatal("refresh errors are reported for the log")
	}
	if _, err := uc.Stats(ctx, "bind"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown provider: %v", err)
	}
}

func TestPause(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{s: map[domain.Provider]domain.Settings{}}
	ag := &fakeSource{stats: domain.Stats{ProtectionEnabled: true}}
	pi := &fakeSource{stats: domain.Stats{ProtectionEnabled: true}}
	log := &audit{}
	uc := NewDNSFilter(Deps{Repo: repo, Providers: map[domain.Provider]DNSFilterProvider{domain.ProviderAdGuard: ag, domain.ProviderPiHole: pi},
		Pausable: []domain.Provider{domain.ProviderPiHole}, Audit: log, Clock: clk{}})
	_, _ = uc.SaveSettings(ctx, SettingsInput{Provider: "pihole", URL: "http://pi.local", Password: "secret"})
	_, _ = uc.SaveSettings(ctx, SettingsInput{Provider: "adguard", URL: "http://ag.local", Username: "a", Password: "secret"})

	if s, _ := uc.Settings(ctx, "pihole"); !s.CanPause {
		t.Fatal("pihole is pausable")
	}
	if s, _ := uc.Settings(ctx, "adguard"); s.CanPause {
		t.Fatal("adguard is not pausable")
	}
	if _, err := uc.Pause(ctx, "adguard", 5); !errors.Is(err, kernel.ErrForbidden) || ag.paused != 0 {
		t.Fatalf("pausing must be opt-in per provider: %v", err)
	}
	if _, err := uc.Pause(ctx, "pihole", 7); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	v, err := uc.Pause(ctx, "pihole", 30)
	if err != nil || pi.paused != 30*time.Minute || v.ProtectionEnabled || v.PausedUntil == nil || !v.PausedUntil.Equal(time.Unix(0, 0).Add(30*time.Minute)) {
		t.Fatalf("%+v %v", v, err)
	}
	if len(log.events) != 1 || log.events[0].Type != "dns_filter_paused" || log.events[0].Detail != "pihole for 30 min" {
		t.Fatalf("security log: %+v", log.events)
	}
	// Paused, but the read-back fails: the widget still shows the pause.
	uc2 := NewDNSFilter(Deps{Repo: repo, Providers: map[domain.Provider]DNSFilterProvider{domain.ProviderPiHole: &pauseOnly{&fakeSource{}}},
		Pausable: []domain.Provider{domain.ProviderPiHole}, Clock: clk{}})
	if v, err := uc2.Pause(ctx, "pihole", 5); err != nil || v.PausedUntil == nil || v.ProtectionEnabled {
		t.Fatalf("%+v %v", v, err)
	}
}

// pauseOnly pauses fine but cannot read statistics.
type pauseOnly struct{ *fakeSource }

func (p *pauseOnly) Stats(context.Context, domain.Settings) (domain.Stats, error) {
	return domain.Stats{}, errors.New("timeout")
}
