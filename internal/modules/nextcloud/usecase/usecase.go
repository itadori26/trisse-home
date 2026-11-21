package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"waylight/internal/modules/nextcloud/domain"
	"waylight/internal/shared/kernel"
)

// SettingsView never contains the app password.
type SettingsView struct {
	Configured  bool
	URL         string
	Username    string
	HasPassword bool
}

// SettingsInput is the input of SaveSettings; an empty app password keeps
// the stored one. SignIn is a finished sign-in of the setup assistant
// (#654) that replaces address, user name and app password.
type SettingsInput struct {
	URL, Username, AppPassword, SignIn string
}

// StatusView is the cached status.
type StatusView struct {
	Status    domain.Status
	UpdatedAt time.Time
	Error     string
}

// connectionTTL is how long the status of a named connection is reused.
// The background job only refreshes the module settings, so connections
// are read on demand at the same pace.
const connectionTTL = 5 * time.Minute

// Nextcloud is the input port of the module.
type Nextcloud interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Status reads through the named connection the widget chose (#610),
	// or through the module settings when connectionID is empty or the
	// connection is not usable.
	Status(ctx context.Context, connectionID string) (StatusView, error)
	Refresh(ctx context.Context) error
}

// Connections resolves a named connection of the integrations module to
// settings (output port, #610). Any error means the module settings are
// used instead.
type Connections interface {
	Settings(ctx context.Context, id string) (domain.Settings, error)
}

// SettingsRepository persists the settings (output port); Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// StatusSource queries Nextcloud (output port).
type StatusSource interface {
	Status(ctx context.Context, s domain.Settings) (domain.Status, error)
}

type nextcloud struct {
	repo    SettingsRepository
	source  StatusSource
	conns   Connections
	clock   kernel.Clock
	signIns kernel.SignIns
	mu      sync.RWMutex
	cached  *StatusView
	// byConnection holds the status per named connection.
	byConnection map[string]StatusView
}

// NewNextcloud returns the Nextcloud use cases; conns (may be nil)
// resolves named connections, signIns (may be nil) hands over sign-ins of
// the setup assistant.
func NewNextcloud(repo SettingsRepository, source StatusSource, conns Connections, clock kernel.Clock, signIns kernel.SignIns) Nextcloud {
	return &nextcloud{repo: repo, source: source, conns: conns, clock: clock, signIns: signIns, byConnection: map[string]StatusView{}}
}

// Settings implements Nextcloud.
func (uc *nextcloud) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: s.AppPassword != ""}, nil
}

// SaveSettings implements Nextcloud; the connection is tested first.
func (uc *nextcloud) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	if in.SignIn != "" {
		g, err := kernel.RedeemSignIn(ctx, uc.signIns, in.SignIn, kernel.SignInNextcloud)
		if err != nil {
			return SettingsView{}, err
		}
		in = SettingsInput{URL: g.URL, Username: g.Username, AppPassword: g.Secret}
	}
	password := in.AppPassword
	if password == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			if kernel.TargetChanged([]string{old.URL, old.Username}, []string{in.URL, in.Username}) {
				return SettingsView{}, kernel.SecretRequired("appPassword")
			}
			password = old.AppPassword
		}
	}
	s, err := domain.NewSettings(in.URL, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	st, err := uc.source.Status(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.SetupFailure("Nextcloud", err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(StatusView{Status: st, UpdatedAt: uc.clock.Now()})
	return SettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: true}, nil
}

// Status implements Nextcloud (cache-aside).
func (uc *nextcloud) Status(ctx context.Context, connectionID string) (StatusView, error) {
	if connectionID != "" && uc.conns != nil {
		if s, err := uc.conns.Settings(ctx, connectionID); err == nil {
			return uc.connectionStatus(ctx, connectionID, s), nil
		}
	}
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()
	if cached != nil {
		return *cached, nil
	}
	if _, err := uc.repo.Load(ctx); err != nil {
		return StatusView{}, err
	}
	_ = uc.Refresh(ctx) // failures are reported through StatusView.Error
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return *uc.cached, nil
}

// Refresh implements Nextcloud; it is run by the background job.
func (uc *nextcloud) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	st, err := uc.source.Status(ctx, s)
	if err != nil {
		uc.mu.Lock()
		view := StatusView{}
		if uc.cached != nil {
			view = *uc.cached
		}
		view.Error = "Nextcloud is not reachable."
		uc.cached = &view
		uc.mu.Unlock()
		return fmt.Errorf("query nextcloud: %w", err)
	}
	uc.store(StatusView{Status: st, UpdatedAt: uc.clock.Now()})
	return nil
}

// connectionStatus returns the cached status of a named connection or
// reads it. Like the module status, a failed read keeps the last status
// and reports the error alongside.
func (uc *nextcloud) connectionStatus(ctx context.Context, id string, s domain.Settings) StatusView {
	now := uc.clock.Now()
	uc.mu.RLock()
	view, ok := uc.byConnection[id]
	uc.mu.RUnlock()
	if ok && now.Sub(view.UpdatedAt) < connectionTTL {
		return view
	}
	if st, err := uc.source.Status(ctx, s); err != nil {
		view.Error = "Nextcloud is not reachable."
	} else {
		view = StatusView{Status: st}
	}
	view.UpdatedAt = now
	uc.mu.Lock()
	uc.byConnection[id] = view
	uc.mu.Unlock()
	return view
}

func (uc *nextcloud) store(v StatusView) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = &v
}
