package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"waylight/internal/modules/media/domain"
	"waylight/internal/shared/kernel"
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

// Media is the input port of the module (#61). Overview and Poster read
// through the named connection the widget chose (#610); an empty
// connectionID, or a connection that is not usable, uses the module
// settings.
type Media interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Overview(ctx context.Context, connectionID string) (domain.Overview, error)
	// Poster returns the cover of a recently added item.
	Poster(ctx context.Context, connectionID, itemID string) (domain.Poster, error)
}

// Connections resolves a named connection of the integrations module to
// settings (output port, #610), whichever media server it is. Any error
// means the module settings are used instead.
type Connections interface {
	Settings(ctx context.Context, id string) (domain.Settings, error)
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

type cachedOverview struct {
	overview domain.Overview
	fetched  time.Time
}

type media struct {
	repo      SettingsRepository
	providers map[domain.Provider]MediaServerProvider
	conns     Connections
	clock     kernel.Clock
	mu        sync.Mutex
	// cache holds one overview per connection ID; "" is the module settings.
	cache map[string]cachedOverview
}

// NewMedia returns the use cases; providers maps each provider to its
// adapter, conns (may be nil) resolves named connections.
func NewMedia(repo SettingsRepository, providers map[domain.Provider]MediaServerProvider, conns Connections, clock kernel.Clock) Media {
	return &media{repo: repo, providers: providers, conns: conns, clock: clock, cache: map[string]cachedOverview{}}
}

// load returns the settings of the named connection, falling back to the
// module settings while the connection does not exist or is not usable.
func (uc *media) load(ctx context.Context, connectionID string) (domain.Settings, error) {
	if connectionID != "" && uc.conns != nil {
		if s, err := uc.conns.Settings(ctx, connectionID); err == nil {
			return s, nil
		}
	}
	return uc.repo.Load(ctx)
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
		return SettingsView{}, kernel.SetupFailure(s.Provider.Title(), err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	clear(uc.cache)
	uc.cache[""] = cachedOverview{overview: o, fetched: uc.clock.Now()}
	uc.mu.Unlock()
	return SettingsView{Configured: true, Provider: string(s.Provider), URL: s.URL, HasToken: true}, nil
}

// Overview implements Media.
func (uc *media) Overview(ctx context.Context, connectionID string) (domain.Overview, error) {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return domain.Overview{}, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.clock.Now()
	if c, ok := uc.cache[connectionID]; ok && now.Sub(c.fetched) < overviewTTL {
		return c.overview, nil
	}
	o, err := uc.providers[s.Provider].Overview(ctx, s)
	if err != nil {
		return domain.Overview{}, kernel.NewPublicError(kernel.ErrUnavailable, "The media server is not reachable.")
	}
	if len(o.Recent) > domain.MaxRecent {
		o.Recent = o.Recent[:domain.MaxRecent]
	}
	uc.cache[connectionID] = cachedOverview{overview: o, fetched: now}
	return o, nil
}

// Poster implements Media. Only IDs of the cached overview are forwarded,
// so the endpoint cannot be used to read arbitrary server paths.
func (uc *media) Poster(ctx context.Context, connectionID, itemID string) (domain.Poster, error) {
	if !domain.ValidItemID(itemID) {
		return domain.Poster{}, domain.ErrPosterNotFound
	}
	o, err := uc.Overview(ctx, connectionID)
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
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return domain.Poster{}, err
	}
	p, err := uc.providers[s.Provider].Poster(ctx, s, itemID)
	if err != nil {
		return domain.Poster{}, domain.ErrPosterNotFound
	}
	return p, nil
}
