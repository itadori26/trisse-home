package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/netip"
	"os"
	"slices"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/adguard"
	"waylight/internal/modules/appearance"
	appearancedomain "waylight/internal/modules/appearance/domain"
	"waylight/internal/modules/attention"
	attentionuc "waylight/internal/modules/attention/usecase"
	"waylight/internal/modules/auth"
	authdomain "waylight/internal/modules/auth/domain"
	"waylight/internal/modules/backup"
	"waylight/internal/modules/backupmonitor"
	"waylight/internal/modules/bookmarks"
	"waylight/internal/modules/briefing"
	"waylight/internal/modules/calendar"
	"waylight/internal/modules/countdown"
	"waylight/internal/modules/dashboard"
	"waylight/internal/modules/discovery"
	discoveryuc "waylight/internal/modules/discovery/usecase"
	"waylight/internal/modules/diskhealth"
	"waylight/internal/modules/docker"
	"waylight/internal/modules/focus"
	"waylight/internal/modules/fritzbox"
	"waylight/internal/modules/integrations"
	integrationsuc "waylight/internal/modules/integrations/usecase"
	"waylight/internal/modules/nextcloud"
	"waylight/internal/modules/notes"
	"waylight/internal/modules/notifications"
	"waylight/internal/modules/onboarding"
	"waylight/internal/modules/parcels"
	"waylight/internal/modules/places"
	placesuc "waylight/internal/modules/places/usecase"
	"waylight/internal/modules/proxmox"
	"waylight/internal/modules/rss"
	"waylight/internal/modules/search"
	"waylight/internal/modules/services"
	servicesusecase "waylight/internal/modules/services/usecase"
	"waylight/internal/modules/setupsignin"
	"waylight/internal/modules/shopping"
	"waylight/internal/modules/smarthome"
	"waylight/internal/modules/speedtest"
	"waylight/internal/modules/system"
	"waylight/internal/modules/timeline"
	"waylight/internal/modules/todos"
	"waylight/internal/modules/transfer"
	"waylight/internal/modules/transit"
	"waylight/internal/modules/waste"
	"waylight/internal/modules/weather"
	"waylight/internal/modules/wiki"
	"waylight/internal/modules/wireguard"
	"waylight/internal/platform/branding"
	"waylight/internal/platform/clock"
	"waylight/internal/platform/config"
	"waylight/internal/platform/crypto"
	"waylight/internal/platform/database"
	"waylight/internal/platform/demo"
	"waylight/internal/platform/httpclient"
	"waylight/internal/platform/idgen"
	"waylight/internal/platform/oidc"
	"waylight/internal/platform/webauthn"
	"waylight/internal/platform/wol"
	"waylight/internal/shared/answers"
	"waylight/internal/shared/kernel"
)

// frontendOnlyWidgetKinds are widgets that need no backend module.
var frontendOnlyWidgetKinds = []string{"clock"}

// healthCheckTimeout bounds a single service status check.
const healthCheckTimeout = 5 * time.Second

// buildModules constructs every enabled module. Auth and the dashboard are
// always enabled: everything else depends on the session guard, and the
// dashboard is the page that hosts all widgets.
func buildModules(cfg config.Config, log *slog.Logger, db *sql.DB, cipher secretStore) (*auth.Module, *modules.Registry, error) {
	// Set again by the ee/ audit log of this server, if it has one (#303).
	changeSink = nil
	brandingSource = nil
	clk := clock.System{}
	ids := idgen.Random{}
	out := newOutbound(cfg, clk, cipher)
	client, secrets := out.client, out.secrets
	registry := &modules.Registry{}
	alerts := &alertRelay{}
	mailer := &mailRelay{}
	scope := &guestScope{registry: registry}

	authModule, err := auth.New(
		auth.Config{
			CookieSecure: cfg.CookieSecure, SessionTTL: cfg.SessionTTL, SetupToken: cfg.SetupToken,
			OnAccountCreated: func(ctx context.Context, profile string) { onAccountCreated(ctx, cfg, log, registry, profile) },
			DemoPassword:     demoPassword(cfg),
			SSO:              ssoConfig(cfg, out.client, clk),
			WebAuthn:         relyingParty(cfg),
			MembersEnabled:   membersEnabled,
			PublicURL:        cfg.PublicURL,
		},
		auth.Deps{DB: db, Clock: clk, Hasher: crypto.NewArgon2id(crypto.DefaultArgon2idParams), Tokens: crypto.RandomTokens{}, IDs: ids, OTP: crypto.TOTP{}, Cipher: secrets, Log: log, Alerts: alerts, GuestScope: scope, Mailer: mailer,
			InstanceName: instanceName},
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
	var notificationsModule *notifications.Module
	if cfg.ModuleEnabled("notifications") {
		n := notifications.New(db, secrets, out.integration, clk, ids, cfg.PublicURL,
			notifications.Push{RelayURL: cfg.PushRelayURL, Sessions: authModule})
		notifier, notificationsModule = n, n
		n.SetInstanceName(instanceName)
		registry.Register(n)
		sinks = append(sinks, n)
		alerts.targets = append(alerts.targets, n)
		mailer.target = n
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

	// Sign-ins of the setup assistant (#654) are redeemed by the modules
	// below, so the module is built first; switched off, they offer
	// pasted keys only.
	var signIns kernel.SignIns
	if cfg.ModuleEnabled("setupsignin") {
		m := setupsignin.New(out.integration, cfg.PublicURL, clk)
		registry.Register(m)
		signIns = m.SignIns()
	}

	// Team calendar (#299): the calendar adds calendars others shared only
	// while ee/ switches it on; without ee/ nothing ever does.
	calendarSharing := calendar.NewSharing()
	// Integrations read the named connection a widget chose; the
	// integrations module is set into it once it is built, last (#610).
	connections := &modules.LateResolver{}
	optional := []struct {
		name  string
		build func() modules.Module
	}{
		{"onboarding", func() modules.Module { return onboarding.New(db, clk) }},
		{"appearance", func() modules.Module {
			m := appearance.New(db, clk, ids)
			m.SetPresetSource(appearancePreset)
			if err := m.SetDefaultTheme(cfg.DefaultTheme); err != nil {
				log.Warn("DEFAULT_THEME is not a built-in theme, the built-in look stays", "theme", cfg.DefaultTheme, "err", err)
			}
			return m
		}},
		{"search", func() modules.Module {
			m := search.New(db)
			authModule.OnAccountDeleted(func(ctx context.Context, accountID string) {
				if err := m.ForgetAccount(ctx, accountID); err != nil {
					log.Warn("search: remove a deleted account's settings", "err", err)
				}
			})
			return m
		}},
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
		{"todos", func() modules.Module { return todos.New(db, clk, ids, alerts, secrets, out.integration, cfg.PublicURL) }},
		{"rss", func() modules.Module { return rss.New(db, client, clk, ids, cfg.SampleFeeds) }},
		{"system", func() modules.Module {
			return system.New(system.Config{ProcPath: cfg.HostProcPath, SysPath: cfg.HostSysPath, DiskPaths: cfg.DiskPaths, Version: version, Client: client, InstallMethod: cfg.InstallMethod, Demo: cfg.DemoMode}, db, sink, clk)
		}},
		{"docker", func() modules.Module { return dockerModule }},
		{"adguard", func() modules.Module {
			return adguard.New(adguard.Config{DB: db, Cipher: secrets, Client: out.integration, Pausable: cfg.DNSPauseProviders, Audit: authModule, Clock: clk, Connections: connections})
		}},
		{"nextcloud", func() modules.Module { return nextcloud.New(db, secrets, out.integration, connections, clk, signIns) }},
		{"parcels", func() modules.Module {
			return parcels.New(db, secrets, client, httpclient.IntegrationDialContext, clk, ids)
		}},
		{"transfer", func() modules.Module {
			return transfer.New(transfer.Deps{Portables: registry.Portables, Configurables: registry.Configurables, Env: os.LookupEnv, Clock: clk, Audit: authModule,
				WidgetKinds: func(kind string) bool {
					return slices.Contains(frontendOnlyWidgetKinds, kind) || slices.Contains(registry.WidgetKinds(), kind)
				}})
		}},
		{"transit", func() modules.Module { return transit.New(client, cfg.TransitAPIURL, clk) }},
		{"shopping", func() modules.Module {
			host := ""
			if cfg.PublicURL != nil {
				host = cfg.PublicURL.Hostname()
			}
			m := shopping.New(db, secrets, out.integration, ids, clk, host, signIns)
			authModule.OnAccountDeleted(func(ctx context.Context, accountID string) {
				if err := m.ForgetAccount(ctx, accountID); err != nil {
					log.Warn("shopping: remove a deleted account's lists", "err", err)
				}
			})
			return m
		}},
		{"waste", func() modules.Module { return waste.New(db, client, clk, time.Local) }},
		{"wireguard", func() modules.Module { return wireguard.New(db, cfg.WireGuardDir, clk) }},
		{"fritzbox", func() modules.Module { return fritzbox.New(db, secrets, out.integration, connections, clk, time.Local) }},
		{"backupmonitor", func() modules.Module {
			return backupmonitor.New(backupmonitor.Deps{DB: db, Tokens: crypto.RandomTokens{}, Clock: clk, IDs: ids, Observations: sink})
		}},
		{"diskhealth", func() modules.Module {
			return diskhealth.New(diskhealth.Config{Dir: cfg.DiskHealthDir, StaleAfter: cfg.DiskHealthStaleAfter}, sink, clk)
		}},
		{"speedtest", func() modules.Module { return speedtest.New(db, secrets, out.integration, connections, sink, clk) }},
		{"calendar", func() modules.Module {
			return calendar.New(db, secrets, out.integration, clk, ids, time.Local, cfg.PublicURL, calendarSharing, signIns)
		}},
		{"proxmox", func() modules.Module {
			return proxmox.New(proxmox.Config{DB: db, Cipher: secrets, Client: out.integration, Pinned: out.pinned,
				ActionGuests: cfg.ProxmoxActionsGuests, Clock: clk, Audit: authModule, Connections: connections})
		}},
		{"smarthome", func() modules.Module {
			return smarthome.New(smarthome.Config{DB: db, Cipher: secrets, Client: out.integration,
				ActionEntities: cfg.SmartHomeActionsEntities, Clock: clk, Audit: authModule, SignIns: signIns, Connections: connections})
		}},
		{"countdown", func() modules.Module { return countdown.New(db, clk, ids) }},
		{"focus", func() modules.Module { return focus.New(db, clk, ids) }},
	}
	for _, m := range optional {
		if cfg.ModuleEnabled(m.name) {
			registry.Register(m.build())
		}
	}
	// Team tabs (#295): the dashboard reads shared tabs only while ee/
	// switches them on; without ee/ nothing ever does.
	tabSharing := dashboard.NewSharing()
	// Modules from cmd/server/module_<name>.go (#193).
	registerSelfRegistered(moduleDeps{cfg: cfg, db: db, clock: clk, ids: ids, out: out, sink: sink, connections: connections,
		alerts: alerts, auth: authModule, registry: registry, wikiSources: &wiki.MemberSources{},
		notifications: notificationsModule, tabSharing: tabSharing, calendarSharing: calendarSharing})
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
	dash := dashboard.New(db, ids, clk, kinds, sample, tabSharing)
	if cfg.DemoMode {
		dash.UseShowcase(sample)
	}
	registry.Register(dash)
	scope.dashboard = dash
	// Places name the widgets that use them, so they need the dashboard (#591).
	if cfg.ModuleEnabled("places") {
		registry.Register(places.New(db, placeWidgets(dash), clk, ids))
	}
	// Connections name the widgets that use them, so they need the dashboard (#610).
	if cfg.ModuleEnabled("integrations") {
		im := integrations.New(integrations.Deps{DB: db, Cipher: secrets, Modules: registry.All(), Widgets: connectionWidgets(dash),
			Managed: func() bool { return managedSections(registry)["integrations"] }, Log: log, Clock: clk, IDs: ids})
		registry.Register(im)
		connections.Set(im)
	}
	// The attention list asks the other modules, so it comes last (#599).
	registry.Register(attention.New(registry.All(), ownWidgets(dash), authModule.MayRead, clk))
	// Discovery asks the other modules how their service is found (#590).
	if cfg.ModuleEnabled("discovery") {
		var found discovery.Containers
		if dockerModule != nil {
			found = discoveryContainers(dockerModule)
		}
		registry.Register(discovery.New(registry.All(), found, out.integration, cfg.AllowedCIDRs, clk))
	}
	wireSmartSearch(registry, authModule.MayRead, dash.WidgetConfig, log, clk)
	return authModule, registry, nil
}

// wireSmartSearch hands the answer sources of the enabled modules to the
// search module (ADR 0026), so no module imports another. A source answers
// only accounts that may read its module's API (members, guests, policies).
func wireSmartSearch(registry *modules.Registry, mayRead func(ctx context.Context, module string) bool,
	widget func(ctx context.Context, kind string) map[string]any, log *slog.Logger, clk kernel.Clock) {
	var sources []answers.Source
	var target *search.Module
	for _, m := range registry.All() {
		if p, ok := m.(answers.Provider); ok {
			sources = append(sources, p.AnswerSources()...)
		}
		if s, ok := m.(*search.Module); ok {
			target = s
		}
	}
	if target != nil {
		target.SetAnswerSources(search.AnswerDeps{Log: log, Clock: clk, MayUse: mayRead, Widget: widget, Sources: sources})
	}
}

// tourRestarter is implemented by the onboarding module.
type tourRestarter interface {
	RestartTour(ctx context.Context) error
}

// onAccountCreated prepares a fresh account: the guided tour is shown again
// and the installation is filled with sample data.
// profileStore and profileSeeder let the composition root apply the
// start-page profile chosen at setup (#281) without the auth module
// knowing about appearance or the dashboard.
type profileStore interface {
	SetProfile(ctx context.Context, profile string) error
}

type profileSeeder interface {
	SeedForProfile(ctx context.Context, profile string) error
}

func onAccountCreated(ctx context.Context, cfg config.Config, log *slog.Logger, registry *modules.Registry, profile string) {
	// The profile is stored even when the client sent none (the web setup no
	// longer asks, #655): an empty one becomes the default, and the owner gets
	// appearance settings of their own instead of reading the system ones.
	if !managedSections(registry)["appearance"] {
		for _, m := range registry.All() {
			if p, ok := m.(profileStore); ok {
				if err := p.SetProfile(ctx, profile); err != nil {
					log.WarnContext(ctx, "could not store the start-page profile", "err", err)
				}
			}
		}
	}
	for _, m := range registry.All() {
		if t, ok := m.(tourRestarter); ok {
			if err := t.RestartTour(ctx); err != nil {
				log.WarnContext(ctx, "could not reset the guided tour", "err", err)
			}
		}
	}
	seedSampleData(ctx, cfg, log, registry, profile)
}

// seedSampleData fills a fresh installation with example data so new users
// immediately see what every module can do. Failures are logged and never
// block the setup.
func seedSampleData(ctx context.Context, cfg config.Config, log *slog.Logger, registry *modules.Registry, profile string) {
	if !cfg.SampleData {
		return
	}
	managed := managedSections(registry)
	for name, s := range registry.Seeders() {
		// Example data would make a section differ from CONFIG_FILE.
		if managed[name] {
			continue
		}
		seed := s.SeedSampleData
		if ps, ok := s.(profileSeeder); ok {
			seed = func(ctx context.Context) error { return ps.SeedForProfile(ctx, profile) }
		}
		if err := seed(ctx); err != nil {
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

// demoPassword is the public password of the demo account, or empty when
// the instance is not a demo.
func demoPassword(cfg config.Config) string {
	if cfg.DemoMode {
		return demo.Password
	}
	return ""
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

// membersEnabled switches on member accounts (#294). It is on since every
// personal-content module stores an owner per record and filters by it
// (ADR 0018); the auth use cases keep the switch so tests can turn members
// off and check the refusal.
var membersEnabled = true

// alertRelay forwards security alerts of the auth module (built first) to
// the notifications module once that exists; without it alerts are dropped.
type alertRelay struct{ targets []kernel.AlertSink }

// Alert implements kernel.AlertSink.
func (a *alertRelay) Alert(ctx context.Context, al kernel.Alert) {
	for _, t := range a.targets {
		t.Alert(ctx, al)
	}
}

// mailRelay forwards account mail of the auth module (built first) to the
// notifications module once that exists; without it nothing can be sent.
type mailRelay struct{ target kernel.Mailer }

// CanSend implements kernel.Mailer.
func (m *mailRelay) CanSend(ctx context.Context) bool {
	return m.target != nil && m.target.CanSend(ctx)
}

// Send implements kernel.Mailer.
func (m *mailRelay) Send(ctx context.Context, mail kernel.Mail) error {
	if m.target == nil {
		return kernel.ErrUnavailable
	}
	return m.target.Send(ctx, mail)
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
// guest may read are those whose widgets are on the guest's tabs; routes a
// module reserves for a widget kind stay closed without it (#168).
type guestScope struct {
	registry  *modules.Registry
	dashboard *dashboard.Module
}

// ReachForTabs implements auth's usecase.GuestScope.
func (g *guestScope) ReachForTabs(ctx context.Context, tabs []string) (authdomain.GuestReach, error) {
	reach := authdomain.GuestReach{Modules: map[string]bool{}}
	var kinds []string
	if g.dashboard != nil && len(tabs) > 0 {
		k, err := g.dashboard.WidgetKindsOnTabs(ctx, tabs)
		if err != nil {
			return authdomain.GuestReach{}, err
		}
		kinds = k
	}
	for _, m := range g.registry.All() {
		for _, k := range m.WidgetKinds() {
			if slices.Contains(kinds, k) {
				reach.Modules[m.Name()] = true
			}
		}
		if gated, ok := m.(modules.GuestGated); ok {
			for prefix, kind := range gated.GuestGatedRoutes() {
				if !slices.Contains(kinds, kind) {
					reach.Closed = append(reach.Closed, prefix)
				}
			}
		}
	}
	slices.Sort(reach.Closed) // map order is random; keep it stable
	return reach, nil
}

// DataKinds implements auth's usecase.GuestScope: the data kind of every
// registered module, which decides what members may change (ADR 0018).
func (g *guestScope) DataKinds() map[string]string {
	out := map[string]string{}
	for _, m := range g.registry.All() {
		out[m.Name()] = string(m.DataKind())
	}
	return out
}

// ssoConfig builds the single sign-on settings of the auth module.
func ssoConfig(cfg config.Config, client *http.Client, clk clock.System) auth.SSOConfig {
	sso := auth.SSOConfig{AllowedSubjects: cfg.OIDCAllowedSubjects, AllowedEmails: cfg.OIDCAllowedEmails,
		ProxyHeader: cfg.AuthProxyHeader, ProxyUsers: cfg.AuthProxyUsers, TrustedProxies: cfg.TrustedProxies,
		PasswordLogin: cfg.PasswordLogin, Require2FA: cfg.SSORequire2FA,
		ButtonLabel: cfg.OIDCButtonLabel, ButtonIcon: cfg.OIDCButtonIcon, AutoRedirect: cfg.OIDCAutoRedirect}
	if cfg.OIDCIssuer != "" {
		sso.OIDC = oidc.New(oidc.Config{Issuer: cfg.OIDCIssuer, ClientID: cfg.OIDCClientID, ClientSecret: cfg.OIDCClientSecret,
			RedirectURL: cfg.OIDCRedirectURL, GroupsClaim: cfg.OIDCGroupsClaim}, client, clk.Now)
	}
	// Sign in with a Microsoft account by code (#535): a public client, no
	// secret, no redirect address.
	// The demo keeps the button to show the feature; the sign-in page answers
	// it with a note instead of contacting Microsoft.
	if cfg.MicrosoftClientID != "" {
		sso.Microsoft = oidc.New(oidc.Config{Issuer: cfg.MicrosoftIssuer, ClientID: cfg.MicrosoftClientID,
			Tenants: cfg.MicrosoftTenants}, client, clk.Now)
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

// instanceName is the instance's name from the branding (#305): the
// organisation's while the ee branding applies, the product's otherwise.
// brandingSource is read on every call, so it may be set later in the build.
func instanceName(ctx context.Context) string {
	b, _ := branding.Current(ctx, brandingSource)
	return b.Name
}

// appearancePreset is the organisation's colour scheme from the branding
// (#305): its default theme and accent colour, and whether members are kept
// on them. None while the ee branding does not apply.
func appearancePreset(ctx context.Context) (appearance.Preset, bool) {
	b, custom := branding.Current(ctx, brandingSource)
	if !custom || b.DefaultTheme == "" {
		return appearance.Preset{}, false
	}
	return appearance.Preset{Theme: appearancedomain.Theme(b.DefaultTheme), Accent: b.Accent, Locked: b.LockTheme}, true
}

// ownWidgets adapts the dashboard's widgets on the account's own tabs to the
// attention list (#599), so neither module imports the other.
func ownWidgets(dash *dashboard.Module) attention.Widgets {
	return func(ctx context.Context) ([]attentionuc.Widget, error) {
		own, err := dash.OwnWidgets(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]attentionuc.Widget, 0, len(own))
		for _, w := range own {
			out = append(out, attentionuc.Widget{ID: w.ID, TabID: w.TabID, Kind: w.Kind, Config: w.Config})
		}
		return out, nil
	}
}

// placeWidgets adapts the dashboard's widgets on the account's own tabs to
// the places module (#591), so neither module imports the other.
func placeWidgets(dash *dashboard.Module) places.Widgets {
	return func(ctx context.Context) ([]placesuc.Widget, error) {
		own, err := dash.OwnWidgets(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]placesuc.Widget, 0, len(own))
		for _, w := range own {
			// Widgets name their place in their settings (#591); settings
			// that are not JSON name none.
			var cfg struct {
				PlaceID string `json:"placeId"`
			}
			_ = json.Unmarshal([]byte(w.Config), &cfg)
			out = append(out, placesuc.Widget{ID: w.ID, TabID: w.TabID, Kind: w.Kind, PlaceID: cfg.PlaceID})
		}
		return out, nil
	}
}

// connectionWidgets adapts the dashboard's widgets on the account's own tabs
// to the integrations module (#610), so neither module imports the other.
func connectionWidgets(dash *dashboard.Module) integrations.Widgets {
	return func(ctx context.Context) ([]integrationsuc.Widget, error) {
		own, err := dash.OwnWidgets(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]integrationsuc.Widget, 0, len(own))
		for _, w := range own {
			// Widgets name their connection in their settings; settings that
			// are not JSON name none.
			var cfg struct {
				ConnectionID string `json:"connectionId"`
			}
			_ = json.Unmarshal([]byte(w.Config), &cfg)
			out = append(out, integrationsuc.Widget{ID: w.ID, TabID: w.TabID, Kind: w.Kind, ConnectionID: cfg.ConnectionID})
		}
		return out, nil
	}
}

// discoveryContainers adapts the docker module's containers to service
// discovery (#590), so neither module imports the other.
func discoveryContainers(d *docker.Module) discovery.Containers {
	return func(ctx context.Context) ([]discoveryuc.Container, error) {
		list, err := d.Containers(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]discoveryuc.Container, 0, len(list))
		for _, c := range list {
			dc := discoveryuc.Container{Name: c.Name, Image: c.Image}
			for _, p := range c.Ports {
				dc.Ports = append(dc.Ports, struct{ Private, Public int }{Private: p.Private, Public: p.Public})
			}
			for _, a := range c.Addresses {
				if addr, err := netip.ParseAddr(a); err == nil {
					dc.Addresses = append(dc.Addresses, addr)
				}
			}
			out = append(out, dc)
		}
		return out, nil
	}
}
