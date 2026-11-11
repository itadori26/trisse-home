package auth

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth/adapter/controller"
	"trisse-home/internal/modules/auth/adapter/gateway"
	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/modules/auth/usecase"
	"trisse-home/internal/shared/kernel"
)

const (
	purgeInterval      = time.Hour
	eventPruneInterval = 24 * time.Hour
)

// Config holds the auth settings derived from the environment.
type Config struct {
	CookieSecure bool
	SessionTTL   time.Duration
	// SetupToken authorises the first setup. If empty, a random token is
	// generated at start and written to the log.
	SetupToken string
	// OnAccountCreated runs after the first-run setup; may be nil.
	OnAccountCreated func(ctx context.Context)
	// Demo enables the password-less demo login (DEMO_MODE).
	Demo bool
}

// observerFunc adapts a function to usecase.SetupObserver.
type observerFunc func(ctx context.Context)

// AccountCreated implements usecase.SetupObserver.
func (f observerFunc) AccountCreated(ctx context.Context) {
	if f != nil {
		f(ctx)
	}
}

// Deps are the infrastructure services the module needs.
type Deps struct {
	DB     *sql.DB
	Clock  kernel.Clock
	Hasher usecase.PasswordHasher
	Tokens usecase.TokenGenerator
	IDs    kernel.IDGenerator
	OTP    usecase.OneTimePasswords
	Cipher gateway.Cipher
	Log    *slog.Logger
	// Alerts receives security alerts (new device, lockouts); may be nil.
	Alerts kernel.AlertSink
	// GuestScope maps a guest's tabs to the modules it may read; may be nil.
	GuestScope usecase.GuestScope
}

// Module wires the auth rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers   *controller.Handlers
	status     usecase.GetStatus
	purge      usecase.PurgeExpiredSessions
	reset      usecase.ResetAccount
	security   usecase.SecurityLog
	setup      usecase.SetupAccount
	tokens     usecase.TokenGenerator
	demo       bool
	setupToken string
	log        *slog.Logger
}

// New builds the auth module.
func New(cfg Config, d Deps) (*Module, error) {
	token := cfg.SetupToken
	if token == "" {
		generated, err := d.Tokens.NewToken()
		if err != nil {
			return nil, fmt.Errorf("generate setup token: %w", err)
		}
		token = generated
	}
	accounts := gateway.NewAccountRepository(d.DB, d.Cipher)
	sessions := gateway.NewSessionRepository(d.DB)
	ttl := usecase.SessionTTL(cfg.SessionTTL)
	events := usecase.NewRecorder(gateway.NewEventRepository(d.DB), d.Alerts, d.IDs, d.Clock, d.Log)
	deps := usecase.Deps{
		Accounts: accounts, Sessions: sessions, Hasher: d.Hasher, Tokens: d.Tokens,
		Limiter: gateway.NewAttemptLimiter(domain.DefaultLoginThrottle), Clock: d.Clock, TTL: ttl, IDs: d.IDs, OTP: d.OTP,
		Events: events,
	}
	securityLog := usecase.NewSecurityLog(events)
	status := usecase.NewGetStatus(accounts, sessions, d.Clock)
	uc := controller.UseCases{
		Status:         status,
		Setup:          usecase.NewSetupAccount(deps, gateway.NewSetupToken(token), observerFunc(cfg.OnAccountCreated)),
		Login:          usecase.NewLogin(deps),
		Logout:         usecase.NewLogout(sessions),
		Authenticate:   usecase.NewAuthenticate(sessions, accounts, d.Clock, ttl),
		Authorizer:     usecase.NewAuthorizer(d.GuestScope),
		Guests:         usecase.NewGuests(deps),
		APITokens:      usecase.NewAPITokens(deps, gateway.NewTokenRepository(d.DB)),
		ChangePassword: usecase.NewChangePassword(deps),
		Sessions:       usecase.NewManageSessions(sessions, d.Clock, events),
		SecurityLog:    securityLog,
		TwoFactor:      usecase.NewTwoFactor(deps),
	}
	if cfg.Demo {
		uc.DemoLogin = usecase.NewDemoLogin(deps)
	}
	return &Module{
		handlers:   controller.NewHandlers(uc, controller.CookieConfig{Secure: cfg.CookieSecure}),
		status:     status,
		purge:      usecase.NewPurgeExpiredSessions(sessions, d.Clock),
		reset:      usecase.NewResetAccount(accounts, sessions),
		security:   securityLog,
		setup:      uc.Setup,
		tokens:     d.Tokens,
		demo:       cfg.Demo,
		setupToken: token,
		log:        d.Log,
	}, nil
}

// Name implements modules.Module.
func (m *Module) Name() string { return "auth" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Public("GET /api/auth/status", m.handlers.Status)
	r.Public("POST /api/auth/setup", m.handlers.Setup)
	r.Public("POST /api/auth/login", m.handlers.Login)
	if m.demo {
		r.Public("POST /api/auth/demo", m.handlers.DemoLogin)
	}
	r.Protected("POST /api/auth/logout", m.handlers.Logout)
	r.Protected("PUT /api/auth/password", m.handlers.ChangePassword)
	r.Protected("GET /api/auth/totp", m.handlers.TwoFactorStatus)
	r.Protected("POST /api/auth/totp/setup", m.handlers.StartTwoFactor)
	r.Protected("POST /api/auth/totp/confirm", m.handlers.ConfirmTwoFactor)
	r.Protected("POST /api/auth/totp/disable", m.handlers.DisableTwoFactor)
	r.Protected("GET /api/auth/sessions", m.handlers.ListSessions)
	r.Protected("GET /api/auth/events", m.handlers.SecurityEvents)
	r.Protected("GET /api/auth/tokens", m.handlers.ListTokens)
	r.Protected("POST /api/auth/tokens", m.handlers.CreateToken)
	r.Protected("DELETE /api/auth/tokens/{id}", m.handlers.RevokeToken)
	r.Protected("GET /api/auth/guests", m.handlers.ListGuests)
	r.Protected("POST /api/auth/guests", m.handlers.CreateGuest)
	r.Protected("PUT /api/auth/guests/{id}", m.handlers.UpdateGuest)
	r.Protected("DELETE /api/auth/guests/{id}", m.handlers.DeleteGuest)
	r.Protected("POST /api/auth/sessions/revoke-others", m.handlers.RevokeOtherSessions)
	r.Protected("DELETE /api/auth/sessions/{id}", m.handlers.RevokeSession)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{
		Name: "auth.purge-expired-sessions", Interval: purgeInterval,
		Run: func(ctx context.Context) error {
			_, err := m.purge.Execute(ctx)
			return err
		},
	}, {
		Name: "auth.prune-security-events", Interval: eventPruneInterval,
		Run: func(ctx context.Context) error {
			_, err := m.security.Prune(ctx)
			return err
		},
	}}
}

// Guard returns the session/CSRF middleware for protected routes.
func (m *Module) Guard(next http.Handler) http.Handler { return m.handlers.Guard(next) }

// AnnounceSetup logs the setup token while no account exists, so that only
// someone with access to the server logs can claim the instance.
func (m *Module) AnnounceSetup(ctx context.Context) error {
	st, err := m.status.Execute(ctx, usecase.StatusRequest{})
	if err != nil {
		return fmt.Errorf("check setup status: %w", err)
	}
	if st.SetupRequired {
		m.log.Warn("no account set up yet: open the web interface and enter this setup token",
			"setup_token", m.setupToken)
	}
	return nil
}

// ResetAccount deletes the account and all sessions (CLI recovery).
func (m *Module) ResetAccount(ctx context.Context) error { return m.reset.Execute(ctx) }

// PrepareDemo creates the demo account with a random, never revealed
// password if none exists yet (which also seeds the sample data).
func (m *Module) PrepareDemo(ctx context.Context) error {
	st, err := m.status.Execute(ctx, usecase.StatusRequest{})
	if err != nil || !st.SetupRequired {
		return err
	}
	password, err := m.tokens.NewToken()
	if err != nil {
		return fmt.Errorf("generate demo password: %w", err)
	}
	if _, err := m.setup.Execute(ctx, usecase.SetupRequest{SetupToken: m.setupToken, Password: password, ClientKey: "demo-setup"}); err != nil {
		return fmt.Errorf("create demo account: %w", err)
	}
	return nil
}

// Audit implements kernel.AuditSink for other modules (e.g. data import).
func (m *Module) Audit(ctx context.Context, e kernel.AuditEvent) { m.security.Audit(ctx, e) }
