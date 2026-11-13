package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"trisse-home/internal/modules/releases/domain"
	"trisse-home/internal/shared/kernel"
)

// ReleasesView is the cached release list.
type ReleasesView struct {
	Releases  []domain.Release
	UpdatedAt time.Time
	// Error lists artists whose releases could not be loaded.
	Error string
}

// Releases is the input port of the module (#63).
type Releases interface {
	SearchArtists(ctx context.Context, query string) ([]domain.Artist, error)
	Followed(ctx context.Context) ([]domain.Artist, error)
	Follow(ctx context.Context, id, name string) error
	Unfollow(ctx context.Context, id string) error
	// List returns the releases of the last weeks, newest first.
	List(ctx context.Context, weeks int) (ReleasesView, error)
	// Refresh reloads the releases of every followed artist (job).
	Refresh(ctx context.Context) error
	Cover(ctx context.Context, releaseID string) (domain.Cover, error)
}

// ArtistRepository persists the followed artists (output port).
type ArtistRepository interface {
	List(ctx context.Context) ([]domain.Artist, error)
	Add(ctx context.Context, a domain.Artist) error
	Remove(ctx context.Context, id string) error
}

// ReleaseProvider reads MusicBrainz and the Cover Art Archive (output port).
type ReleaseProvider interface {
	SearchArtists(ctx context.Context, query string) ([]domain.Artist, error)
	// ReleasesSince returns the release groups of an artist first released on or after since.
	ReleasesSince(ctx context.Context, artist domain.Artist, since time.Time) ([]domain.Release, error)
	Cover(ctx context.Context, releaseID string) (domain.Cover, error)
}

type releases struct {
	repo     ArtistRepository
	provider ReleaseProvider
	clock    kernel.Clock
	mu       sync.RWMutex
	cached   *ReleasesView
}

// NewReleases returns the use cases.
func NewReleases(repo ArtistRepository, provider ReleaseProvider, clock kernel.Clock) Releases {
	return &releases{repo: repo, provider: provider, clock: clock}
}

// SearchArtists implements Releases.
func (uc *releases) SearchArtists(ctx context.Context, query string) ([]domain.Artist, error) {
	query = strings.TrimSpace(query)
	if utf8.RuneCountInString(query) < domain.MinQueryLength {
		return []domain.Artist{}, nil
	}
	list, err := uc.provider.SearchArtists(ctx, query)
	if err != nil {
		return nil, kernel.NewPublicError(kernel.ErrUnavailable, "MusicBrainz is not reachable.")
	}
	return list, nil
}

// Followed implements Releases.
func (uc *releases) Followed(ctx context.Context) ([]domain.Artist, error) {
	list, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list artists: %w", err)
	}
	return list, nil
}

// Follow implements Releases; the new artist's releases are loaded at once.
func (uc *releases) Follow(ctx context.Context, id, name string) error {
	a, err := domain.NewArtist(id, name)
	if err != nil {
		return err
	}
	list, err := uc.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list artists: %w", err)
	}
	if slices.ContainsFunc(list, func(x domain.Artist) bool { return x.ID == a.ID }) {
		return nil
	}
	if len(list) >= domain.MaxFollowed {
		return domain.ErrTooManyArtists
	}
	if err := uc.repo.Add(ctx, a); err != nil {
		return fmt.Errorf("add artist: %w", err)
	}
	uc.merge(ctx, a)
	return nil
}

// Unfollow implements Releases.
func (uc *releases) Unfollow(ctx context.Context, id string) error {
	list, err := uc.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list artists: %w", err)
	}
	i := slices.IndexFunc(list, func(x domain.Artist) bool { return x.ID == id })
	if i < 0 {
		return nil
	}
	if err := uc.repo.Remove(ctx, id); err != nil {
		return fmt.Errorf("remove artist: %w", err)
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.cached != nil {
		name := list[i].Name
		uc.cached.Releases = slices.DeleteFunc(slices.Clone(uc.cached.Releases), func(r domain.Release) bool { return r.Artist == name })
	}
	return nil
}

// List implements Releases (cache-aside; the job keeps it fresh).
func (uc *releases) List(ctx context.Context, weeks int) (ReleasesView, error) {
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()
	if cached == nil {
		if err := uc.Refresh(ctx); err != nil && !errors.Is(err, errPartial) {
			return ReleasesView{}, err
		}
		uc.mu.RLock()
		cached = uc.cached
		uc.mu.RUnlock()
	}
	v := *cached
	v.Releases = domain.Recent(v.Releases, weeks, uc.clock.Now())
	return v, nil
}

var errPartial = errors.New("some artists could not be loaded")

// Refresh implements Releases. Requests are sequential because MusicBrainz
// allows one per second; the provider enforces the rate.
func (uc *releases) Refresh(ctx context.Context) error {
	artists, err := uc.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list artists: %w", err)
	}
	now := uc.clock.Now()
	since := now.AddDate(0, 0, -7*domain.MaxWeeks)
	v := ReleasesView{UpdatedAt: now, Releases: []domain.Release{}}
	var failed []string
	for _, a := range artists {
		list, err := uc.provider.ReleasesSince(ctx, a, since)
		if err != nil {
			failed = append(failed, a.Name)
			continue
		}
		v.Releases = append(v.Releases, list...)
	}
	if len(failed) > 0 {
		v.Error = "Releases could not be loaded for: " + strings.Join(failed, ", ")
	}
	uc.mu.Lock()
	uc.cached = &v
	uc.mu.Unlock()
	if len(failed) > 0 {
		return fmt.Errorf("%w: %s", errPartial, strings.Join(failed, ", "))
	}
	return nil
}

// merge adds the releases of a newly followed artist to the cache.
func (uc *releases) merge(ctx context.Context, a domain.Artist) {
	list, err := uc.provider.ReleasesSince(ctx, a, uc.clock.Now().AddDate(0, 0, -7*domain.MaxWeeks))
	if err != nil {
		return // the next refresh tries again
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.cached != nil {
		uc.cached.Releases = append(slices.Clone(uc.cached.Releases), list...)
	}
}

// Cover implements Releases. Only covers of listed releases are proxied.
func (uc *releases) Cover(ctx context.Context, releaseID string) (domain.Cover, error) {
	if !domain.ValidMBID(releaseID) {
		return domain.Cover{}, domain.ErrCoverNotFound
	}
	uc.mu.RLock()
	known := uc.cached != nil && slices.ContainsFunc(uc.cached.Releases, func(r domain.Release) bool { return r.ID == releaseID })
	uc.mu.RUnlock()
	if !known {
		return domain.Cover{}, domain.ErrCoverNotFound
	}
	c, err := uc.provider.Cover(ctx, releaseID)
	if err != nil {
		return domain.Cover{}, domain.ErrCoverNotFound
	}
	return c, nil
}
