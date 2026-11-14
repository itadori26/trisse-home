package usecase

import (
	"context"
	"slices"
	"sync"
	"time"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/shared/kernel"
)

// healthTTL: the DWD publishes pollen once a day (11:00) and UV once a day
// (07:30); an hourly look picks up the new issue soon enough.
const healthTTL = time.Hour

// HealthOptions are the selectable pollen regions and UV cities.
type HealthOptions struct {
	Regions []domain.PollenRegion
	Cities  []string
}

// HealthReport is the pollen and UV forecast of one widget. A part is nil
// when it was not asked for or is not available.
type HealthReport struct {
	Pollen *domain.PollenForecast
	UV     *domain.UVForecast
}

// Health is the input port of the pollen and UV forecasts (#56).
type Health interface {
	Options(ctx context.Context) (HealthOptions, error)
	// Report returns the pollen forecast of region (0 = none) filtered to
	// types (empty = all) and the UV forecast of city ("" = none).
	Report(ctx context.Context, region int, city string, types []domain.PollenType) (HealthReport, error)
}

// HealthProvider loads the DWD health forecasts of all regions (output port).
type HealthProvider interface {
	Pollen(ctx context.Context) ([]domain.PollenForecast, error)
	UV(ctx context.Context) ([]domain.UVForecast, error)
}

// ErrHealthUnavailable is returned when a forecast was never loaded.
var ErrHealthUnavailable = kernel.NewPublicError(kernel.ErrUnavailable, "The pollen and UV forecasts are not available right now.")

type cachedFeed[T any] struct {
	list []T
	at   time.Time
}

type health struct {
	provider HealthProvider
	clock    kernel.Clock

	mu     sync.Mutex
	pollen cachedFeed[domain.PollenForecast]
	uv     cachedFeed[domain.UVForecast]
}

// NewHealth returns the pollen and UV use cases. Nothing is requested
// until a widget asks for a region or city.
func NewHealth(p HealthProvider, clock kernel.Clock) Health {
	return &health{provider: p, clock: clock}
}

// load returns the cached feed or loads it; stale data is preferred over an error.
func load[T any](ctx context.Context, uc *health, c *cachedFeed[T], fetch func(context.Context) ([]T, error)) ([]T, error) {
	now := uc.clock.Now()
	uc.mu.Lock()
	cur := *c
	uc.mu.Unlock()
	if !cur.at.IsZero() && now.Sub(cur.at) < healthTTL {
		return cur.list, nil
	}
	list, err := fetch(ctx)
	if err != nil {
		if !cur.at.IsZero() {
			return cur.list, nil
		}
		return nil, ErrHealthUnavailable
	}
	uc.mu.Lock()
	*c = cachedFeed[T]{list: list, at: now}
	uc.mu.Unlock()
	return list, nil
}

func (uc *health) Options(ctx context.Context) (HealthOptions, error) {
	pollen, err := load(ctx, uc, &uc.pollen, uc.provider.Pollen)
	if err != nil {
		return HealthOptions{}, err
	}
	uv, err := load(ctx, uc, &uc.uv, uc.provider.UV)
	if err != nil {
		return HealthOptions{}, err
	}
	o := HealthOptions{Regions: make([]domain.PollenRegion, 0, len(pollen)), Cities: make([]string, 0, len(uv))}
	for _, f := range pollen {
		o.Regions = append(o.Regions, f.Region)
	}
	for _, f := range uv {
		o.Cities = append(o.Cities, f.City)
	}
	slices.Sort(o.Cities)
	return o, nil
}

func (uc *health) Report(ctx context.Context, region int, city string, types []domain.PollenType) (HealthReport, error) {
	var r HealthReport
	if region != 0 {
		pollen, err := load(ctx, uc, &uc.pollen, uc.provider.Pollen)
		if err != nil {
			return HealthReport{}, err
		}
		if i := slices.IndexFunc(pollen, func(f domain.PollenForecast) bool { return f.Region.ID == region }); i >= 0 {
			f := pollen[i]
			f.Days = f.Of(types)
			r.Pollen = &f
		}
	}
	if city != "" {
		uv, err := load(ctx, uc, &uc.uv, uc.provider.UV)
		if err != nil {
			return HealthReport{}, err
		}
		if i := slices.IndexFunc(uv, func(f domain.UVForecast) bool { return f.City == city }); i >= 0 {
			r.UV = &uv[i]
		}
	}
	return r, nil
}
