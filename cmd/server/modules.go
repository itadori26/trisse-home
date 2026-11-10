package main

import (
	"database/sql"
	"log/slog"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/appearance"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/modules/dashboard"
	"trisse-home/internal/modules/search"
	"trisse-home/internal/modules/services"
	"trisse-home/internal/modules/weather"
	"trisse-home/internal/platform/clock"
	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/crypto"
	"trisse-home/internal/platform/httpclient"
	"trisse-home/internal/platform/idgen"
)

// frontendOnlyWidgetKinds are widgets that need no backend module.
var frontendOnlyWidgetKinds = []string{"clock"}

// healthCheckTimeout bounds a single service status check.
const healthCheckTimeout = 5 * time.Second

// buildModules constructs every enabled module. Auth and the dashboard are
// always enabled: everything else depends on the session guard, and the
// dashboard is the page that hosts all widgets.
func buildModules(cfg config.Config, log *slog.Logger, db *sql.DB, secrets *crypto.AESGCM) (*auth.Module, *modules.Registry, error) {
	_ = secrets
	clk := clock.System{}
	ids := idgen.Random{}
	client := httpclient.New()
	registry := &modules.Registry{}

	authModule, err := auth.New(
		auth.Config{CookieSecure: cfg.CookieSecure, SessionTTL: cfg.SessionTTL, SetupToken: cfg.SetupToken},
		auth.Deps{DB: db, Clock: clk, Hasher: crypto.NewArgon2id(crypto.DefaultArgon2idParams), Tokens: crypto.RandomTokens{}, Log: log},
	)
	if err != nil {
		return nil, nil, err
	}
	registry.Register(authModule)

	optional := []struct {
		name  string
		build func() modules.Module
	}{
		{"appearance", func() modules.Module { return appearance.New(db) }},
		{"search", func() modules.Module { return search.New(db) }},
		{"weather", func() modules.Module { return weather.New(client, clk) }},
		{"services", func() modules.Module {
			return services.New(services.Deps{
				DB: db, Clock: clk, IDs: ids,
				Client:         httpclient.New(httpclient.WithTimeout(healthCheckTimeout)),
				InsecureClient: httpclient.New(httpclient.WithTimeout(healthCheckTimeout), httpclient.WithInsecureTLS()),
			})
		}},
	}
	for _, m := range optional {
		if cfg.ModuleEnabled(m.name) {
			registry.Register(m.build())
		}
	}

	kinds := append(append([]string{}, frontendOnlyWidgetKinds...), registry.WidgetKinds()...)
	registry.Register(dashboard.New(db, ids, kinds))
	return authModule, registry, nil
}
