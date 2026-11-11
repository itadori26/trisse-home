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
	"trisse-home/internal/modules/auth/usecase"
	"trisse-home/internal/shared/kernel"
)

const (
	maxFailedAttempts    = 5
	failedAttemptsWindow = 15 * time.Minute
	purgeInterval        = time.Hour
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
	Log    *slog.Logger
}

// Module wires the auth rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers   *controller.Handlers
	status     usecase.GetStatus
	purge      usecase.PurgeExpiredSessions
	reset      usecase.ResetAccount
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
	accounts := gateway.NewAccountRepository(d.DB)
	sessions := gateway.NewSessionRepository(d.DB)
	ttl := usecase.SessionTTL(cfg.SessionTTL)
	deps := usecase.Deps{
		Accounts: accounts, Sessions: sessions, Hasher: d.Hasher, Tokens: d.Tokens,
		Limiter: gateway.NewAttemptLimiter(maxFailedAttempts, failedAttemptsWindow), Clock: d.Clock, TTL: ttl, IDs: d.IDs,
	}
	status := usecase.NewGetStatus(accounts, sessions, d.Clock)
	uc := controller.UseCases{
		Status:         status,
		Setup:          usecase.NewSetupAccount(deps, gateway.NewSetupToken(token), observerFunc(cfg.OnAccountCreated)),
		Login:          usecase.NewLogin(deps),
		Logout:         usecase.NewLogout(sessions),
		Authenticate:   usecase.NewAuthenticate(sessions, d.Clock, ttl),
		ChangePassword: usecase.NewChangePassword(deps),
		Sessions:       usecase.NewManageSessions(sessions, d.Clock),
	}
	return &Module{
		handlers:   controller.NewHandlers(uc, controller.CookieConfig{Secure: cfg.CookieSecure}),
		status:     status,
		purge:      usecase.NewPurgeExpiredSessions(sessions, d.Clock),
		reset:      usecase.NewResetAccount(accounts, sessions),
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
	r.Protected("POST /api/auth/logout", m.handlers.Logout)
	r.Protected("PUT /api/auth/password", m.handlers.ChangePassword)
	r.Protected("GET /api/auth/sessions", m.handlers.ListSessions)
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
