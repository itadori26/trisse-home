package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/shared/kernel"
)

// alertsTTL: DWD updates warnings and radar every five minutes.
const alertsTTL = 5 * time.Minute

// Radar image size requested from the provider.
const (
	RadarWidth  = 640
	RadarHeight = 520
)

// RadarImage is a PNG with its capture time.
type RadarImage struct {
	PNG       []byte
	FetchedAt time.Time
}

// Alerts is the input port for official warnings and the rain radar.
type Alerts interface {
	Warnings(ctx context.Context, lat, lon float64) ([]domain.Warning, error)
	Radar(ctx context.Context, lat, lon float64) (RadarImage, error)
}

// WarningProvider loads official warnings (output port).
type WarningProvider interface {
	Warnings(ctx context.Context, loc domain.Location) ([]domain.Warning, error)
}

// RadarProvider renders the rain radar for an area (output port).
type RadarProvider interface {
	Radar(ctx context.Context, box domain.BBox, width, height int) ([]byte, error)
}

// ErrRadarCoverage is returned outside the DWD radar coverage.
var ErrRadarCoverage = kernel.NewPublicError(kernel.ErrInvalid, "The rain radar is only available for Germany.")

type cachedWarnings struct {
	list []domain.Warning
	at   time.Time
}

type alerts struct {
	warnings WarningProvider
	radar    RadarProvider
	clock    kernel.Clock

	mu            sync.Mutex
	warningsCache map[string]cachedWarnings
	radarCache    map[string]RadarImage
}

// NewAlerts returns the warning and radar use cases.
func NewAlerts(w WarningProvider, r RadarProvider, clock kernel.Clock) Alerts {
	return &alerts{warnings: w, radar: r, clock: clock, warningsCache: map[string]cachedWarnings{}, radarCache: map[string]RadarImage{}}
}

// Warnings implements Alerts. Outside Germany the list is empty; cached
// data is preferred over an error.
func (uc *alerts) Warnings(ctx context.Context, lat, lon float64) ([]domain.Warning, error) {
	loc, err := domain.NewLocation(lat, lon)
	if err != nil {
		return nil, err
	}
	if !loc.InGermany() {
		return []domain.Warning{}, nil
	}
	now := uc.clock.Now()
	uc.mu.Lock()
	c, ok := uc.warningsCache[loc.Key()]
	uc.mu.Unlock()
	if !ok || now.Sub(c.at) >= alertsTTL {
		list, err := uc.warnings.Warnings(ctx, loc)
		if err != nil {
			if ok {
				return domain.RelevantWarnings(c.list, now), nil
			}
			return nil, kernel.NewPublicError(kernel.ErrUnavailable, "The weather warnings are not available right now.")
		}
		c = cachedWarnings{list: list, at: now}
		uc.mu.Lock()
		uc.warningsCache[loc.Key()] = c
		uc.mu.Unlock()
	}
	return domain.RelevantWarnings(c.list, now), nil
}

// Radar implements Alerts.
func (uc *alerts) Radar(ctx context.Context, lat, lon float64) (RadarImage, error) {
	loc, err := domain.NewLocation(lat, lon)
	if err != nil {
		return RadarImage{}, err
	}
	if !loc.InGermany() {
		return RadarImage{}, ErrRadarCoverage
	}
	now := uc.clock.Now()
	uc.mu.Lock()
	img, ok := uc.radarCache[loc.Key()]
	uc.mu.Unlock()
	if ok && now.Sub(img.FetchedAt) < alertsTTL {
		return img, nil
	}
	png, err := uc.radar.Radar(ctx, loc.RadarBBox(), RadarWidth, RadarHeight)
	if err != nil {
		if ok {
			return img, nil
		}
		return RadarImage{}, fmt.Errorf("load radar: %w", kernel.ErrUnavailable)
	}
	img = RadarImage{PNG: png, FetchedAt: now}
	uc.mu.Lock()
	uc.radarCache[loc.Key()] = img
	uc.mu.Unlock()
	return img, nil
}
