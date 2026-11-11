package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/adguard"
	"trisse-home/internal/modules/appearance"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/modules/backup"
	"trisse-home/internal/modules/backupmonitor"
	"trisse-home/internal/modules/bookmarks"
	"trisse-home/internal/modules/calendar"
	"trisse-home/internal/modules/dashboard"
	"trisse-home/internal/modules/docker"
	"trisse-home/internal/modules/fritzbox"
	"trisse-home/internal/modules/nextcloud"
	"trisse-home/internal/modules/notes"
	"trisse-home/internal/modules/notifications"
	"trisse-home/internal/modules/onboarding"
	"trisse-home/internal/modules/parcels"
	"trisse-home/internal/modules/rss"
	"trisse-home/internal/modules/search"
	"trisse-home/internal/modules/services"
	"trisse-home/internal/modules/shopping"
	"trisse-home/internal/modules/system"
	"trisse-home/internal/modules/todos"
	"trisse-home/internal/modules/transfer"
	"trisse-home/internal/modules/transit"
	"trisse-home/internal/modules/waste"
	"trisse-home/internal/modules/weather"
	"trisse-home/internal/modules/wireguard"
	"trisse-home/internal/platform/clock"
	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/crypto"
	"trisse-home/internal/platform/demo"
	"trisse-home/internal/platform/httpclient"
	"trisse-home/internal/platform/idgen"
	"trisse-home/internal/platform/wol"
	"trisse-home/internal/shared/kernel"
)

// frontendOnlyWidgetKinds are widgets that need no backend module.
var frontendOnlyWidgetKinds = []string{"clock"}

// healthCheckTimeout bounds a single service status check.
const healthCheckTimeout = 5 * time.Second

// buildModules constructs every enabled module. Auth and the dashboard are
// always enabled: everything else depends on the session guard, and the
// dashboard is the page that hosts all widgets.
func buildModules(cfg config.Config, log *slog.Logger, db *sql.DB, cipher secretStore) (*auth.Module, *modules.Registry, error) {
	clk := clock.System{}
	ids := idgen.Random{}
	out := newOutbound(cfg, clk, cipher)
	client, secrets := out.client, out.secrets
	registry := &modules.Registry{}
	alerts := &alertRelay{}

	authModule, err := auth.New(
		auth.Config{
			CookieSecure: cfg.CookieSecure, SessionTTL: cfg.SessionTTL, SetupToken: cfg.SetupToken,
			OnAccountCreated: func(ctx context.Context) { onAccountCreated(ctx, cfg, log, registry) },
			Demo:             cfg.DemoMode,
		},
		auth.Deps{DB: db, Clock: clk, Hasher: crypto.NewArgon2id(crypto.DefaultArgon2idParams), Tokens: crypto.RandomTokens{}, IDs: ids, OTP: crypto.TOTP{}, Cipher: secrets, Log: log, Alerts: alerts},
	)
	if err != nil {
		return nil, nil, err
	}
	registry.Register(authModule)

	// Notifications are built first: monitoring modules report observations
	// to them, or to a sink that discards them when the module is disabled.
	var sink kernel.ObservationSink = kernel.DiscardObservations{}
	if cfg.ModuleEnabled("notifications") {
		n := notifications.New(db, secrets, client, clk)
		registry.Register(n)
		sink = n
		alerts.target = n
	}

	optional := []struct {
		name  string
		build func() modules.Module
	}{
		{"onboarding", func() modules.Module { return onboarding.New(db, clk) }},
		{"appearance", func() modules.Module { return appearance.New(db) }},
		{"search", func() modules.Module { return search.New(db) }},
		{"weather", func() modules.Module { return weather.New(client, clk) }},
		{"services", func() modules.Module {
			return services.New(services.Deps{
				DB: db, Clock: clk, IDs: ids, Observations: sink,
				Waker: out.waker, Client: out.health, InsecureClient: out.insecureHealth,
			})
		}},
		{"bookmarks", func() modules.Module { return bookmarks.New(db, clk, ids) }},
		{"notes", func() modules.Module { return notes.New(db, clk, ids) }},
		{"todos", func() modules.Module { return todos.New(db, clk, ids) }},
		{"rss", func() modules.Module { return rss.New(db, client, clk, ids, cfg.SampleFeeds) }},
		{"system", func() modules.Module {
			return system.New(system.Config{ProcPath: cfg.HostProcPath, SysPath: cfg.HostSysPath, DiskPaths: cfg.DiskPaths, Version: version, Client: client}, db, sink, clk)
		}},
		{"docker", func() modules.Module {
			return docker.New(client, docker.Config{ProxyURL: cfg.DockerProxyURL, Actions: cfg.DockerActions,
				ActionsContainers: cfg.DockerActionsContainers, ActionsProxyURL: cfg.DockerActionsProxyURL}, sink, clk)
		}},
		{"adguard", func() modules.Module { return adguard.New(db, secrets, client, clk) }},
		{"nextcloud", func() modules.Module { return nextcloud.New(db, secrets, client, clk) }},
		{"parcels", func() modules.Module { return parcels.New(db, secrets, client, clk, ids) }},
		{"transfer", func() modules.Module { return transfer.New(registry.Portables, clk, authModule) }},
		{"transit", func() modules.Module { return transit.New(client, cfg.TransitAPIURL, clk) }},
		{"shopping", func() modules.Module { return shopping.New(db, secrets, client, ids) }},
		{"waste", func() modules.Module { return waste.New(db, client, clk, time.Local) }},
		{"wireguard", func() modules.Module { return wireguard.New(db, cfg.WireGuardDir, clk) }},
		{"fritzbox", func() modules.Module { return fritzbox.New(db, secrets, client, clk, time.Local) }},
		{"backupmonitor", func() modules.Module {
			return backupmonitor.New(backupmonitor.Deps{DB: db, Tokens: crypto.RandomTokens{}, Clock: clk, IDs: ids, Observations: sink})
		}},
		{"calendar", func() modules.Module { return calendar.New(db, secrets, client, clk, ids, time.Local) }},
		// scaffold:modules - make new-module inserts new modules above this line.
	}
	for _, m := range optional {
		if cfg.ModuleEnabled(m.name) {
			registry.Register(m.build())
		}
	}
	if cfg.ModuleEnabled("backup") && !cfg.DemoMode {
		b, err := backup.New(backup.Config{Dir: cfg.BackupDir, Interval: cfg.BackupInterval, Retention: cfg.BackupRetention}, db, clk)
		if err != nil {
			return nil, nil, err
		}
		registry.Register(b)
	}

	kinds := append(append([]string{}, frontendOnlyWidgetKinds...), registry.WidgetKinds()...)
	sample := dashboard.SampleWeather{Label: cfg.SampleWeather.Name, Lat: cfg.SampleWeather.Lat, Lon: cfg.SampleWeather.Lon}
	registry.Register(dashboard.New(db, ids, kinds, sample))
	return authModule, registry, nil
}

// tourRestarter is implemented by the onboarding module.
type tourRestarter interface {
	RestartTour(ctx context.Context) error
}

// onAccountCreated prepares a fresh account: the guided tour is shown again
// and the installation is filled with sample data.
func onAccountCreated(ctx context.Context, cfg config.Config, log *slog.Logger, registry *modules.Registry) {
	for _, m := range registry.All() {
		if t, ok := m.(tourRestarter); ok {
			if err := t.RestartTour(ctx); err != nil {
				log.WarnContext(ctx, "could not reset the guided tour", "err", err)
			}
		}
	}
	seedSampleData(ctx, cfg, log, registry)
}

// seedSampleData fills a fresh installation with example data so new users
// immediately see what every module can do. Failures are logged and never
// block the setup.
func seedSampleData(ctx context.Context, cfg config.Config, log *slog.Logger, registry *modules.Registry) {
	if !cfg.SampleData {
		return
	}
	for name, s := range registry.Seeders() {
		if err := s.SeedSampleData(ctx); err != nil {
			log.WarnContext(ctx, "could not create sample data", "module", name, "err", err)
		}
	}
	log.InfoContext(ctx, "sample data created", "modules", len(registry.Seeders()))
}

// secretStore encrypts credentials before modules store them.
type secretStore interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// outbound bundles the adapters that reach other systems. In demo mode they
// are replaced by fixtures, so no module ever contacts a real system.
type outbound struct {
	client, health, insecureHealth *http.Client
	waker                          interface {
		Wake(context.Context, kernel.MAC) error
	}
	secrets secretStore
}

func newOutbound(cfg config.Config, clk clock.System, cipher secretStore) outbound {
	if cfg.DemoMode {
		return outbound{
			client:         &http.Client{Transport: demo.Transport{Now: clk.Now}, Timeout: healthCheckTimeout},
			health:         &http.Client{Transport: demo.HealthyTransport{}},
			insecureHealth: &http.Client{Transport: demo.HealthyTransport{}},
			waker:          demo.NoWake{},
			secrets:        demo.ReadOnlySecrets{Inner: cipher},
		}
	}
	return outbound{
		client:         httpclient.New(),
		health:         httpclient.New(httpclient.WithTimeout(healthCheckTimeout)),
		insecureHealth: httpclient.New(httpclient.WithTimeout(healthCheckTimeout), httpclient.WithInsecureTLS()),
		waker:          wol.Sender{Relay: cfg.WolRelay, Broadcast: cfg.WolBroadcast},
		secrets:        cipher,
	}
}

// alertRelay forwards security alerts of the auth module (built first) to
// the notifications module once that exists; without it alerts are dropped.
type alertRelay struct{ target kernel.AlertSink }

// Alert implements kernel.AlertSink.
func (a *alertRelay) Alert(ctx context.Context, al kernel.Alert) {
	if a.target != nil {
		a.target.Alert(ctx, al)
	}
}
