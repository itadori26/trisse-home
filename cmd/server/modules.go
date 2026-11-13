package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/adguard"
	"trisse-home/internal/modules/appearance"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/modules/backup"
	"trisse-home/internal/modules/backupmonitor"
	"trisse-home/internal/modules/bookmarks"
	"trisse-home/internal/modules/briefing"
	"trisse-home/internal/modules/calendar"
	"trisse-home/internal/modules/countdown"
	"trisse-home/internal/modules/dashboard"
	"trisse-home/internal/modules/diskhealth"
	"trisse-home/internal/modules/docker"
	"trisse-home/internal/modules/focus"
	"trisse-home/internal/modules/fritzbox"
	"trisse-home/internal/modules/nextcloud"
	"trisse-home/internal/modules/notes"
	"trisse-home/internal/modules/notifications"
	"trisse-home/internal/modules/onboarding"
	"trisse-home/internal/modules/parcels"
	"trisse-home/internal/modules/proxmox"
	"trisse-home/internal/modules/rss"
	"trisse-home/internal/modules/search"
	"trisse-home/internal/modules/services"
	servicesusecase "trisse-home/internal/modules/services/usecase"
	"trisse-home/internal/modules/shopping"
	"trisse-home/internal/modules/smarthome"
	"trisse-home/internal/modules/speedtest"
	"trisse-home/internal/modules/system"
	"trisse-home/internal/modules/timeline"
	"trisse-home/internal/modules/todos"
	"trisse-home/internal/modules/transfer"
	"trisse-home/internal/modules/transit"
	"trisse-home/internal/modules/waste"
	"trisse-home/internal/modules/weather"
	"trisse-home/internal/modules/wireguard"
	"trisse-home/internal/platform/clock"
	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/crypto"
	"trisse-home/internal/platform/database"
	"trisse-home/internal/platform/demo"
	"trisse-home/internal/platform/httpclient"
	"trisse-home/internal/platform/idgen"
	"trisse-home/internal/platform/oidc"
	"trisse-home/internal/platform/webauthn"
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
	scope := &guestScope{registry: registry}

	authModule, err := auth.New(
		auth.Config{
			CookieSecure: cfg.CookieSecure, SessionTTL: cfg.SessionTTL, SetupToken: cfg.SetupToken,
			OnAccountCreated: func(ctx context.Context) { onAccountCreated(ctx, cfg, log, registry) },
			Demo:             cfg.DemoMode,
			SSO:              ssoConfig(cfg, out.client, clk),
			WebAuthn:         relyingParty(cfg),
		},
		auth.Deps{DB: db, Clock: clk, Hasher: crypto.NewArgon2id(crypto.DefaultArgon2idParams), Tokens: crypto.RandomTokens{}, IDs: ids, OTP: crypto.TOTP{}, Cipher: secrets, Log: log, Alerts: alerts, GuestScope: scope},
	)
	if err != nil {
		return nil, nil, err
	}
	registry.Register(authModule)

	// Notifications are built first: monitoring modules report observations
	// to them, or to a sink that discards them when the module is disabled.
	// The timeline records the same observations and alerts (#74).
	sinks := observationFanOut{}
	var maintenance maintenanceReader
	var notifier kernel.AlertSink
	if cfg.ModuleEnabled("notifications") {
		n := notifications.New(db, secrets, out.integration, clk, ids)
		notifier = n
		registry.Register(n)
		sinks = append(sinks, n)
		alerts.targets = append(alerts.targets, n)
		maintenance = n
	}
	if cfg.ModuleEnabled("timeline") {
		tl := timeline.New(db, clk, ids)
		registry.Register(tl)
		sinks = append(sinks, tl)
		alerts.targets = append(alerts.targets, tl)
	}
	var sink kernel.ObservationSink = sinks

	// The docker module is built ahead of the list: the services module
	// reads its containers for autodiscovery (#78).
	var dockerModule *docker.Module
	var containers servicesusecase.ContainerSource
	if cfg.ModuleEnabled("docker") {
		dockerModule = docker.New(client, docker.Config{ProxyURL: cfg.DockerProxyURL, Actions: cfg.DockerActions,
			ActionsContainers: cfg.DockerActionsContainers, ActionsProxyURL: cfg.DockerActionsProxyURL,
			DB: db, Cipher: secrets, Registry: client}, sink, clk)
		containers = containerLabels{docker: dockerModule}
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
				Waker: out.waker, Client: out.health, InsecureClient: out.insecureHealth, Maintenance: maintenance,
				Containers: containers,
			})
		}},
		{"bookmarks", func() modules.Module { return bookmarks.New(db, clk, ids, out.guarded) }},
		{"notes", func() modules.Module { return notes.New(db, clk, ids) }},
		{"todos", func() modules.Module { return todos.New(db, clk, ids, alerts) }},
		{"rss", func() modules.Module { return rss.New(db, client, clk, ids, cfg.SampleFeeds) }},
		{"system", func() modules.Module {
			return system.New(system.Config{ProcPath: cfg.HostProcPath, SysPath: cfg.HostSysPath, DiskPaths: cfg.DiskPaths, Version: version, Client: client, InstallMethod: cfg.InstallMethod}, db, sink, clk)
		}},
		{"docker", func() modules.Module { return dockerModule }},
		{"adguard", func() modules.Module { return adguard.New(db, secrets, out.integration, clk) }},
		{"nextcloud", func() modules.Module { return nextcloud.New(db, secrets, out.integration, clk) }},
		{"parcels", func() modules.Module { return parcels.New(db, secrets, client, clk, ids) }},
		{"transfer", func() modules.Module { return transfer.New(registry.Portables, clk, authModule) }},
		{"transit", func() modules.Module { return transit.New(client, cfg.TransitAPIURL, clk) }},
		{"shopping", func() modules.Module { return shopping.New(db, secrets, out.integration, ids) }},
		{"waste", func() modules.Module { return waste.New(db, client, clk, time.Local) }},
		{"wireguard", func() modules.Module { return wireguard.New(db, cfg.WireGuardDir, clk) }},
		{"fritzbox", func() modules.Module { return fritzbox.New(db, secrets, out.integration, clk, time.Local) }},
		{"backupmonitor", func() modules.Module {
			return backupmonitor.New(backupmonitor.Deps{DB: db, Tokens: crypto.RandomTokens{}, Clock: clk, IDs: ids, Observations: sink})
		}},
		{"diskhealth", func() modules.Module {
			return diskhealth.New(diskhealth.Config{Dir: cfg.DiskHealthDir, StaleAfter: cfg.DiskHealthStaleAfter}, sink, clk)
		}},
		{"speedtest", func() modules.Module { return speedtest.New(db, secrets, out.integration, sink, clk) }},
		{"calendar", func() modules.Module { return calendar.New(db, secrets, out.integration, clk, ids, time.Local) }},
		{"proxmox", func() modules.Module {
			return proxmox.New(proxmox.Config{DB: db, Cipher: secrets, Client: out.integration, Pinned: out.pinned,
				ActionGuests: cfg.ProxmoxActionsGuests, Clock: clk, Audit: authModule})
		}},
		{"smarthome", func() modules.Module {
			return smarthome.New(smarthome.Config{DB: db, Cipher: secrets, Client: out.integration,
				ActionEntities: cfg.SmartHomeActionsEntities, Clock: clk, Audit: authModule})
		}},
		{"countdown", func() modules.Module { return countdown.New(db, clk, ids) }},
		{"focus", func() modules.Module { return focus.New(db, clk, ids) }},
		// scaffold:modules - make new-module inserts new modules above this line.
	}
	for _, m := range optional {
		if cfg.ModuleEnabled(m.name) {
			registry.Register(m.build())
		}
	}
	if cfg.ModuleEnabled("backup") && !cfg.DemoMode {
		b, err := backup.New(backup.Config{Dir: cfg.BackupDir, Interval: cfg.BackupInterval, Retention: cfg.BackupRetention,
			Cipher: secrets, Client: out.backupRemote, Alerts: alerts, Log: log,
			Swapper: database.NewRestorer(cfg.DatabasePath, backup.MaxRestoreBytes), Reauth: authModule, Restart: requestRestart}, db, clk)
		if err != nil {
			return nil, nil, err
		}
		registry.Register(b)
	}

	// The briefing reads other modules, so it is built after them (#52).
	if cfg.ModuleEnabled("briefing") {
		bc := briefing.Config{Sources: briefingSources(registry), DefaultLat: cfg.SampleWeather.Lat, DefaultLon: cfg.SampleWeather.Lon}
		if notifier != nil {
			bc.Publisher = briefingPublisher{notifier}
		}
		registry.Register(briefing.New(db, clk, bc))
	}

	kinds := append(append([]string{}, frontendOnlyWidgetKinds...), registry.WidgetKinds()...)
	sample := dashboard.SampleWeather{Label: cfg.SampleWeather.Name, Lat: cfg.SampleWeather.Lat, Lon: cfg.SampleWeather.Lon}
	dash := dashboard.New(db, ids, clk, kinds, sample)
	registry.Register(dash)
	scope.dashboard = dash
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
	// guarded fetches user-supplied URLs (bookmark titles, favicons, link
	// checks, #72); nil in demo mode, which never contacts other sites.
	guarded *http.Client
	waker   interface {
		Wake(context.Context, kernel.MAC) error
	}
	secrets secretStore
	// integration calls owner-configured targets (AdGuard, Nextcloud,
	// Proxmox, …) with the address guard, no redirects and a response limit
	// (#191); backupRemote is the same with room for whole backups.
	integration, backupRemote *http.Client
	// pinned builds clients that trust one certificate fingerprint (Proxmox).
	pinned func(fingerprint string) *http.Client
}

// bookmarkFetchTimeout bounds title, favicon and link check requests.
const bookmarkFetchTimeout = 8 * time.Second

func newOutbound(cfg config.Config, clk clock.System, cipher secretStore) outbound {
	if cfg.DemoMode {
		return outbound{
			client:         &http.Client{Transport: demo.Transport{Now: clk.Now}, Timeout: healthCheckTimeout},
			health:         &http.Client{Transport: demo.HealthyTransport{}},
			insecureHealth: &http.Client{Transport: demo.HealthyTransport{}},
			waker:          demo.NoWake{},
			secrets:        demo.ReadOnlySecrets{Inner: cipher},
			integration:    &http.Client{Transport: demo.Transport{Now: clk.Now}, Timeout: healthCheckTimeout},
			backupRemote:   &http.Client{Transport: demo.Transport{Now: clk.Now}, Timeout: healthCheckTimeout},
			pinned: func(string) *http.Client {
				return &http.Client{Transport: demo.Transport{Now: clk.Now}, Timeout: healthCheckTimeout}
			},
		}
	}
	return outbound{
		client:         httpclient.New(),
		health:         httpclient.New(httpclient.WithTimeout(healthCheckTimeout)),
		insecureHealth: httpclient.New(httpclient.WithTimeout(healthCheckTimeout), httpclient.WithInsecureTLS()),
		guarded:        httpclient.New(httpclient.WithTimeout(bookmarkFetchTimeout), httpclient.WithInternalAddressGuard()),
		waker:          wol.Sender{Relay: cfg.WolRelay, Broadcast: cfg.WolBroadcast},
		secrets:        cipher,
		integration:    httpclient.New(httpclient.WithIntegrationGuard()),
		backupRemote: httpclient.New(httpclient.WithIntegrationGuard(),
			httpclient.WithMaxResponseBytes(backup.MaxRestoreBytes)),
		pinned: func(fp string) *http.Client {
			return httpclient.New(httpclient.WithPinnedCertificate(fp), httpclient.WithIntegrationGuard())
		},
	}
}

// alertRelay forwards security alerts of the auth module (built first) to
// the notifications module once that exists; without it alerts are dropped.
type alertRelay struct{ targets []kernel.AlertSink }

// Alert implements kernel.AlertSink.
func (a *alertRelay) Alert(ctx context.Context, al kernel.Alert) {
	for _, t := range a.targets {
		t.Alert(ctx, al)
	}
}

// observationFanOut sends every observation to all consumers
// (notifications, timeline); empty when both are disabled.
type observationFanOut []kernel.ObservationSink

// Observe implements kernel.ObservationSink.
func (f observationFanOut) Observe(ctx context.Context, o kernel.Observation) {
	for _, s := range f {
		s.Observe(ctx, o)
	}
}

// guestScope implements the auth module's GuestScope port: the modules a
// guest may read are those whose widgets are on the guest's tabs.
type guestScope struct {
	registry  *modules.Registry
	dashboard *dashboard.Module
}

// ModulesForTabs implements auth's usecase.GuestScope.
func (g *guestScope) ModulesForTabs(ctx context.Context, tabs []string) (map[string]bool, error) {
	out := map[string]bool{}
	if g.dashboard == nil || len(tabs) == 0 {
		return out, nil
	}
	kinds, err := g.dashboard.WidgetKindsOnTabs(ctx, tabs)
	if err != nil {
		return nil, err
	}
	for _, m := range g.registry.All() {
		for _, k := range m.WidgetKinds() {
			if slices.Contains(kinds, k) {
				out[m.Name()] = true
			}
		}
	}
	return out, nil
}

// ssoConfig builds the single sign-on settings of the auth module.
func ssoConfig(cfg config.Config, client *http.Client, clk clock.System) auth.SSOConfig {
	sso := auth.SSOConfig{AllowedSubjects: cfg.OIDCAllowedSubjects, AllowedEmails: cfg.OIDCAllowedEmails,
		ProxyHeader: cfg.AuthProxyHeader, ProxyUsers: cfg.AuthProxyUsers, TrustedProxies: cfg.TrustedProxies,
		PasswordLogin: cfg.PasswordLogin, Require2FA: cfg.SSORequire2FA}
	if cfg.OIDCIssuer != "" {
		sso.OIDC = oidc.New(oidc.Config{Issuer: cfg.OIDCIssuer, ClientID: cfg.OIDCClientID, ClientSecret: cfg.OIDCClientSecret,
			RedirectURL: cfg.OIDCRedirectURL}, client, clk.Now)
	}
	return sso
}

// relyingParty returns the WebAuthn relying party; passkeys need PUBLIC_URL.
func relyingParty(cfg config.Config) *webauthn.RelyingParty {
	if cfg.PublicURL == nil {
		return nil
	}
	return &webauthn.RelyingParty{ID: cfg.PublicURL.Hostname(), Origin: cfg.PublicURL.Scheme + "://" + cfg.PublicURL.Host}
}

// maintenanceReader is implemented by the notifications module (maintenance
// windows, #69) and used by the services module to exclude planned
// downtime from the uptime (#81).
type maintenanceReader interface {
	UnderMaintenance(ctx context.Context, kind, key, label string, at time.Time) bool
}
