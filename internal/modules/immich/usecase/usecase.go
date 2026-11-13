package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"trisse-home/internal/modules/immich/domain"
	"trisse-home/internal/shared/kernel"
)

// overviewTTL bounds how often Immich is asked; statistics change slowly
// and are fetched on demand only.
const overviewTTL = 5 * time.Minute

// SettingsView never contains the API key.
type SettingsView struct {
	Configured bool
	URL        string
	HasAPIKey  bool
}

// SettingsInput is the input of SaveSettings; an empty key keeps the stored one.
type SettingsInput struct {
	URL, APIKey string
}

// Immich is the input port of the module (#85).
type Immich interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Overview(ctx context.Context) (domain.Overview, error)
	// Thumbnail returns the preview of a photo of today's memories.
	Thumbnail(ctx context.Context, assetID string) (domain.Thumbnail, error)
}

// SettingsRepository persists the settings (output port). Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Client talks to the Immich API (output port).
type Client interface {
	Version(ctx context.Context, s domain.Settings) (string, error)
	Stats(ctx context.Context, s domain.Settings) (domain.LibraryStats, error)
	// Memories returns today's memories; an older Immich without the
	// endpoint returns none.
	Memories(ctx context.Context, s domain.Settings, day time.Time) ([]domain.Memory, error)
	Thumbnail(ctx context.Context, s domain.Settings, assetID string) (domain.Thumbnail, error)
}

type immich struct {
	repo    SettingsRepository
	client  Client
	clock   kernel.Clock
	mu      sync.Mutex
	cached  *domain.Overview
	fetched time.Time
}

// NewImmich returns the use cases.
func NewImmich(repo SettingsRepository, client Client, clock kernel.Clock) Immich {
	return &immich{repo: repo, client: client, clock: clock}
}

// Settings implements Immich.
func (uc *immich) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, URL: s.URL, HasAPIKey: true}, nil
}

// SaveSettings implements Immich; the key is tested with the statistics
// call, which needs the asset.statistics permission.
func (uc *immich) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	key := in.APIKey
	if key == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			key = old.APIKey
		}
	}
	s, err := domain.NewSettings(in.URL, key)
	if err != nil {
		return SettingsView{}, err
	}
	if _, err := uc.client.Stats(ctx, s); err != nil {
		return SettingsView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not connect to Immich: {detail}", "detail", err.Error())
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	uc.cached = nil
	uc.mu.Unlock()
	return SettingsView{Configured: true, URL: s.URL, HasAPIKey: true}, nil
}

// Overview implements Immich. Memories and version are optional: their
// errors (e.g. a key without memory.read) only leave them empty.
func (uc *immich) Overview(ctx context.Context) (domain.Overview, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Overview{}, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.clock.Now()
	if uc.cached != nil && now.Sub(uc.fetched) < overviewTTL && sameDay(uc.fetched, now) {
		return *uc.cached, nil
	}
	stats, err := uc.client.Stats(ctx, s)
	if err != nil {
		return domain.Overview{}, kernel.NewPublicError(kernel.ErrUnavailable, "Immich is not reachable.")
	}
	o := domain.Overview{Stats: stats}
	o.Version, _ = uc.client.Version(ctx, s)       // optional
	memories, _ := uc.client.Memories(ctx, s, now) // optional, needs memory.read
	for _, m := range memories {
		if len(m.AssetIDs) > domain.MaxMemoryAssets {
			m.AssetIDs = m.AssetIDs[:domain.MaxMemoryAssets]
		}
		o.Memories = append(o.Memories, m)
	}
	uc.cached, uc.fetched = &o, now
	return o, nil
}

// Thumbnail implements Immich. Only assets of today's memories are
// proxied, so the endpoint cannot be used to browse the library.
func (uc *immich) Thumbnail(ctx context.Context, assetID string) (domain.Thumbnail, error) {
	if !domain.ValidAssetID(assetID) {
		return domain.Thumbnail{}, domain.ErrThumbnailNotFound
	}
	o, err := uc.Overview(ctx)
	if err != nil {
		return domain.Thumbnail{}, err
	}
	known := slices.ContainsFunc(o.Memories, func(m domain.Memory) bool { return slices.Contains(m.AssetIDs, assetID) })
	if !known {
		return domain.Thumbnail{}, domain.ErrThumbnailNotFound
	}
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Thumbnail{}, err
	}
	t, err := uc.client.Thumbnail(ctx, s, assetID)
	if err != nil {
		return domain.Thumbnail{}, domain.ErrThumbnailNotFound
	}
	return t, nil
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
