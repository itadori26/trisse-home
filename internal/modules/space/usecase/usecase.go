package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/space/domain"
	"trisse-home/internal/shared/kernel"
)

// Space is the input port of the module.
type Space interface {
	Settings(ctx context.Context) (domain.Settings, error)
	SaveSettings(ctx context.Context, launches, picture bool, key string, keepKey bool) (domain.Settings, error)
	Launches(ctx context.Context, n int) ([]domain.Launch, error)
	Picture(ctx context.Context) (domain.Picture, error)
	// PictureImage returns today's image (proxied, so the browser does not
	// contact NASA); a video day returns ErrNoImage.
	PictureImage(ctx context.Context) (Image, error)
}

// Image is an image with its media type.
type Image struct {
	Bytes       []byte
	ContentType string
}

// ErrNoImage is returned on days whose picture is a video.
var ErrNoImage = kernel.NewPublicError(kernel.ErrNotFound, "Today's picture is not an image.")

// ErrUnavailable is returned when a source failed and nothing is cached.
var ErrUnavailable = kernel.NewPublicError(kernel.ErrUnavailable, "The space data is not available right now.")

// SettingsRepository stores the settings with the encrypted key (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// LaunchProvider loads the next launches (output port).
type LaunchProvider interface {
	Upcoming(ctx context.Context, limit int) ([]domain.Launch, error)
}

// PictureProvider loads the picture of the day and its image (output port).
type PictureProvider interface {
	Today(ctx context.Context, apiKey string) (domain.Picture, error)
	Image(ctx context.Context, url string) (Image, error)
}

// Cache lifetimes: Launch Library 2 allows 15 requests an hour without a
// key; the picture changes once a day.
const (
	launchTTL  = time.Hour
	pictureTTL = 3 * time.Hour
)

type space struct {
	repo     SettingsRepository
	launches LaunchProvider
	pictures PictureProvider
	clock    kernel.Clock

	mu         sync.Mutex
	launchList []domain.Launch
	launchAt   time.Time
	picture    domain.Picture
	pictureAt  time.Time
	image      Image
	imageURL   string
}

// New returns the use cases.
func New(repo SettingsRepository, launches LaunchProvider, pictures PictureProvider, clock kernel.Clock) Space {
	return &space{repo: repo, launches: launches, pictures: pictures, clock: clock}
}

func (uc *space) Settings(ctx context.Context) (domain.Settings, error) { return uc.repo.Load(ctx) }

func (uc *space) SaveSettings(ctx context.Context, launches, picture bool, key string, keepKey bool) (domain.Settings, error) {
	stored, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Settings{}, err
	}
	s, err := domain.NewSettings(launches, picture, key, keepKey, stored)
	if err != nil {
		return domain.Settings{}, err
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return domain.Settings{}, fmt.Errorf("save settings: %w", err)
	}
	if s.NASAKey != stored.NASAKey {
		uc.mu.Lock()
		uc.pictureAt = time.Time{} // a new key may fix a rate-limited picture
		uc.mu.Unlock()
	}
	return s, nil
}

func (uc *space) Launches(ctx context.Context, n int) ([]domain.Launch, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !s.Launches {
		return nil, domain.ErrDisabled
	}
	n = min(max(n, 1), domain.MaxLaunches)
	now := uc.clock.Now()
	uc.mu.Lock()
	list, at := uc.launchList, uc.launchAt
	uc.mu.Unlock()
	if at.IsZero() || now.Sub(at) >= launchTTL {
		fresh, err := uc.launches.Upcoming(ctx, domain.MaxLaunches)
		switch {
		case err == nil:
			list = fresh
			uc.mu.Lock()
			uc.launchList, uc.launchAt = fresh, now
			uc.mu.Unlock()
		case at.IsZero():
			return nil, ErrUnavailable
		}
	}
	// Launches that took off more than an hour ago drop out of a cached list.
	out := make([]domain.Launch, 0, n)
	for _, l := range list {
		if l.Net.After(now.Add(-time.Hour)) && len(out) < n {
			out = append(out, l)
		}
	}
	return out, nil
}

func (uc *space) Picture(ctx context.Context) (domain.Picture, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Picture{}, err
	}
	if !s.Picture {
		return domain.Picture{}, domain.ErrDisabled
	}
	now := uc.clock.Now()
	uc.mu.Lock()
	p, at := uc.picture, uc.pictureAt
	uc.mu.Unlock()
	if !at.IsZero() && now.Sub(at) < pictureTTL {
		return p, nil
	}
	fresh, err := uc.pictures.Today(ctx, s.NASAKey)
	if err != nil {
		if !at.IsZero() {
			return p, nil
		}
		return domain.Picture{}, ErrUnavailable
	}
	uc.mu.Lock()
	uc.picture, uc.pictureAt = fresh, now
	uc.mu.Unlock()
	return fresh, nil
}

func (uc *space) PictureImage(ctx context.Context) (Image, error) {
	p, err := uc.Picture(ctx)
	if err != nil {
		return Image{}, err
	}
	if !p.IsImage() {
		return Image{}, ErrNoImage
	}
	uc.mu.Lock()
	img, url := uc.image, uc.imageURL
	uc.mu.Unlock()
	if url == p.URL && len(img.Bytes) > 0 {
		return img, nil
	}
	img, err = uc.pictures.Image(ctx, p.URL)
	if err != nil {
		if errors.Is(err, kernel.ErrInvalid) {
			return Image{}, err
		}
		return Image{}, ErrUnavailable
	}
	uc.mu.Lock()
	uc.image, uc.imageURL = img, p.URL
	uc.mu.Unlock()
	return img, nil
}
