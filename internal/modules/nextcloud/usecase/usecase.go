package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/nextcloud/domain"
	"trisse-home/internal/shared/kernel"
)

// SettingsView never contains the app password.
type SettingsView struct {
	Configured  bool
	URL         string
	Username    string
	HasPassword bool
}

// SettingsInput is the input of SaveSettings; an empty app password keeps
// the stored one.
type SettingsInput struct {
	URL, Username, AppPassword string
}

// StatusView is the cached status.
type StatusView struct {
	Status    domain.Status
	UpdatedAt time.Time
	Error     string
}

// Nextcloud is the input port of the module.
type Nextcloud interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Status(ctx context.Context) (StatusView, error)
	Refresh(ctx context.Context) error
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
	repo   SettingsRepository
	source StatusSource
	clock  kernel.Clock
	mu     sync.RWMutex
	cached *StatusView
}

// NewNextcloud returns the Nextcloud use cases.
func NewNextcloud(repo SettingsRepository, source StatusSource, clock kernel.Clock) Nextcloud {
	return &nextcloud{repo: repo, source: source, clock: clock}
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
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(StatusView{Status: st, UpdatedAt: uc.clock.Now()})
	return SettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: true}, nil
}

// Status implements Nextcloud (cache-aside).
func (uc *nextcloud) Status(ctx context.Context) (StatusView, error) {
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

func (uc *nextcloud) store(v StatusView) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = &v
}
