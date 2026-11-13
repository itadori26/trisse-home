package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/media/domain"
	"trisse-home/internal/shared/kernel"
)

// overviewTTL bounds how often the media server is asked; the overview is
// fetched on demand only, so nothing is polled while no widget is open.
const overviewTTL = 15 * time.Second

// SettingsView never contains the token.
type SettingsView struct {
	Configured bool
	Provider   string
	URL        string
	HasToken   bool
}

// SettingsInput is the input of SaveSettings; an empty token keeps the
// stored one.
type SettingsInput struct {
	Provider, URL, Token string
}

// Media is the input port of the module (#61).
type Media interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Overview(ctx context.Context) (domain.Overview, error)
	// Poster returns the cover of a recently added item.
	Poster(ctx context.Context, itemID string) (domain.Poster, error)
}

// SettingsRepository persists the settings (output port). Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// MediaServerProvider reads from one kind of media server (output port);
// Jellyfin and Plex are adapters.
type MediaServerProvider interface {
	Overview(ctx context.Context, s domain.Settings) (domain.Overview, error)
	Poster(ctx context.Context, s domain.Settings, itemID string) (domain.Poster, error)
}

type media struct {
	repo      SettingsRepository
	providers map[domain.Provider]MediaServerProvider
	clock     kernel.Clock
	mu        sync.Mutex
	cached    *domain.Overview
	fetched   time.Time
}

// NewMedia returns the use cases; providers maps each provider to its adapter.
func NewMedia(repo SettingsRepository, providers map[domain.Provider]MediaServerProvider, clock kernel.Clock) Media {
	return &media{repo: repo, providers: providers, clock: clock}
}

// Settings implements Media.
func (uc *media) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{Provider: string(domain.ProviderJellyfin)}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, Provider: string(s.Provider), URL: s.URL, HasToken: true}, nil
}

// SaveSettings implements Media; the connection is tested first.
func (uc *media) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	token := in.Token
	if token == "" {
		if old, err := uc.repo.Load(ctx); err == nil && string(old.Provider) == in.Provider {
			token = old.Token
		}
	}
	s, err := domain.NewSettings(in.Provider, in.URL, token)
	if err != nil {
		return SettingsView{}, err
	}
	o, err := uc.providers[s.Provider].Overview(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not connect to the media server: {detail}", "detail", err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	uc.cached, uc.fetched = &o, uc.clock.Now()
	uc.mu.Unlock()
	return SettingsView{Configured: true, Provider: string(s.Provider), URL: s.URL, HasToken: true}, nil
}

// Overview implements Media.
func (uc *media) Overview(ctx context.Context) (domain.Overview, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Overview{}, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.clock.Now()
	if uc.cached != nil && now.Sub(uc.fetched) < overviewTTL {
		return *uc.cached, nil
	}
	o, err := uc.providers[s.Provider].Overview(ctx, s)
	if err != nil {
		return domain.Overview{}, kernel.NewPublicError(kernel.ErrUnavailable, "The media server is not reachable.")
	}
	if len(o.Recent) > domain.MaxRecent {
		o.Recent = o.Recent[:domain.MaxRecent]
	}
	uc.cached, uc.fetched = &o, now
	return o, nil
}

// Poster implements Media. Only IDs of the cached overview are forwarded,
// so the endpoint cannot be used to read arbitrary server paths.
func (uc *media) Poster(ctx context.Context, itemID string) (domain.Poster, error) {
	if !domain.ValidItemID(itemID) {
		return domain.Poster{}, domain.ErrPosterNotFound
	}
	o, err := uc.Overview(ctx)
	if err != nil {
		return domain.Poster{}, err
	}
	known := false
	for _, it := range o.Recent {
		known = known || it.ID == itemID
	}
	if !known {
		return domain.Poster{}, domain.ErrPosterNotFound
	}
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Poster{}, err
	}
	p, err := uc.providers[s.Provider].Poster(ctx, s, itemID)
	if err != nil {
		return domain.Poster{}, domain.ErrPosterNotFound
	}
	return p, nil
}
