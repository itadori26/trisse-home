package main

import (
	"database/sql"
	"log/slog"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/platform/clock"
	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/crypto"
)

// frontendOnlyWidgetKinds are widgets that need no backend module.
var frontendOnlyWidgetKinds = []string{"clock"}

// buildModules constructs every enabled module. Auth is always enabled
// because all other modules depend on its session guard.
func buildModules(cfg config.Config, log *slog.Logger, db *sql.DB, secrets *crypto.AESGCM) (*auth.Module, *modules.Registry, error) {
	_ = secrets
	clk := clock.System{}
	registry := &modules.Registry{}

	authModule, err := auth.New(
		auth.Config{CookieSecure: cfg.CookieSecure, SessionTTL: cfg.SessionTTL, SetupToken: cfg.SetupToken},
		auth.Deps{DB: db, Clock: clk, Hasher: crypto.NewArgon2id(crypto.DefaultArgon2idParams), Tokens: crypto.RandomTokens{}, Log: log},
	)
	if err != nil {
		return nil, nil, err
	}
	registry.Register(authModule)
	return authModule, registry, nil
}
