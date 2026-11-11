package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // the distroless image has no zoneinfo; calendars need time zones

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/crypto"
	"trisse-home/internal/platform/database"
	"trisse-home/internal/platform/httpserver"
	"trisse-home/internal/platform/logging"
	"trisse-home/internal/platform/middleware"
	"trisse-home/internal/platform/router"
	"trisse-home/internal/platform/scheduler"
	"trisse-home/internal/platform/spa"
	"trisse-home/internal/platform/webui"
	"trisse-home/internal/platform/wol"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

const healthcheckTimeout = 3 * time.Second

func main() {
	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	var err error
	switch cmd {
	case "serve":
		err = serve()
	case "healthcheck":
		err = healthcheck()
	case "wol-relay":
		err = wolRelay()
	case "reset-account":
		err = resetAccount()
	case "version":
		fmt.Println(version)
	default:
		err = fmt.Errorf("unknown command %q (available: serve, healthcheck, wol-relay, reset-account, version)", cmd)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// app holds everything built from the configuration.
type app struct {
	cfg      config.Config
	log      *slog.Logger
	db       *sql.DB
	auth     *auth.Module
	registry *modules.Registry
}

// loadConfig reads the configuration and sets up the default logger.
func loadConfig() (config.Config, *slog.Logger, error) {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return config.Config{}, nil, fmt.Errorf("invalid configuration:\n%w", err)
	}
	log := logging.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)
	return cfg, log, nil
}

// build opens the database, constructs all enabled modules and applies their migrations.
func build(ctx context.Context, cfg config.Config, log *slog.Logger) (*app, error) {
	db, err := database.Open(ctx, cfg.DatabasePath)
	if err != nil {
		return nil, err
	}
	secrets, err := crypto.NewAESGCM(cfg.SecretKey)
	if err != nil {
		return nil, err
	}
	authModule, registry, err := buildModules(cfg, log, db, secrets)
	if err != nil {
		return nil, err
	}
	for _, m := range registry.All() {
		if err := database.Migrate(ctx, db, m.Name(), m.Migrations()); err != nil {
			return nil, err
		}
	}
	return &app{cfg: cfg, log: log, db: db, auth: authModule, registry: registry}, nil
}

func serve() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, log, err := loadConfig()
	if err != nil {
		return err
	}
	a, err := build(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer a.db.Close()
	a.log.Info("starting trisse-home", "version", version, "modules", a.registry.Names())
	if err := a.auth.AnnounceSetup(ctx); err != nil {
		return err
	}
	sched := newScheduler(a)
	sched.Start(ctx)
	defer sched.Wait()
	return httpserver.Run(ctx, a.cfg.ListenAddr, newHandler(a), a.cfg.ShutdownTimeout, a.log)
}

// newHandler builds the router with all module routes and the middleware chain.
func newHandler(a *app) http.Handler {
	r := router.New(a.auth.Guard)
	for _, m := range a.registry.All() {
		m.RegisterRoutes(r)
	}
	r.Protected("GET /api/modules", a.registry.Handler(frontendOnlyWidgetKinds))
	r.Handle("GET /healthz", router.HealthCheck(a.db.PingContext))
	r.Handle("/", spa.Handler(webui.Dist()))

	return middleware.Chain(r,
		middleware.Recover(a.log),
		middleware.ClientIP(a.cfg.TrustedProxies),
		middleware.RequestLog(a.log),
		middleware.AllowList(a.cfg.AllowedCIDRs, a.log),
		middleware.SecurityHeaders,
		http.NewCrossOriginProtection().Handler,
	)
}

// newScheduler registers the background jobs of all modules.
func newScheduler(a *app) *scheduler.Scheduler {
	sched := scheduler.New(a.log)
	for _, m := range a.registry.All() {
		for _, j := range m.Jobs() {
			sched.Add(scheduler.Task{Name: j.Name, Interval: j.Interval, Run: j.Run})
		}
	}
	return sched
}

func resetAccount() error {
	ctx := context.Background()
	cfg, log, err := loadConfig()
	if err != nil {
		return err
	}
	a, err := build(ctx, cfg, log)
	if err != nil {
		return err
	}
	defer a.db.Close()
	if err := a.auth.ResetAccount(ctx); err != nil {
		return err
	}
	fmt.Println("Account removed. Restart the server and use the setup token from the log to set a new password.")
	return nil
}

// healthcheck is used by the container HEALTHCHECK, because the distroless
// image has neither curl nor wget.
func healthcheck() error {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8090"
	}
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("parse LISTEN_ADDR: %w", err)
	}
	client := &http.Client{Timeout: healthcheckTimeout}
	res, err := client.Get("http://127.0.0.1:" + port + "/healthz") //nolint:gosec,noctx // fixed loopback URL, short-lived CLI call
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New("unhealthy: " + res.Status)
	}
	return nil
}

// wolRelay runs the Wake-on-LAN relay (compose profile "wol"). It only reads
// WOL_RELAY_LISTEN and WOL_BROADCAST and needs no database or secret key.
func wolRelay() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	env := func(key, def string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return def
	}
	return wol.Relay{
		Listen:    env("WOL_RELAY_LISTEN", ":"+wol.DefaultRelayPort),
		Broadcast: env("WOL_BROADCAST", wol.DefaultBroadcast),
		Log:       slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		Burst:     10,
		Every:     time.Second,
	}.Serve(ctx)
}
