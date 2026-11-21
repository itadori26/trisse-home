package usecase

import (
	"context"
	"strings"
	"sync"
	"time"

	"waylight/internal/modules/transit/domain"
	"waylight/internal/shared/kernel"
)

// cacheTTL protects the (community-run) provider: one request per stop and
// minute at most.
const cacheTTL = time.Minute

// Transit is the input port of the module.
type Transit interface {
	SearchStops(ctx context.Context, query string) ([]domain.Stop, error)
	// StopsNear lists the stops closest to a place (#694).
	StopsNear(ctx context.Context, lat, lon float64) ([]domain.Stop, error)
	Departures(ctx context.Context, stopID string, n int) ([]domain.Departure, error)
}

// Provider queries a journey planner (output port).
type Provider interface {
	SearchStops(ctx context.Context, query string) ([]domain.Stop, error)
	StopsNear(ctx context.Context, lat, lon float64) ([]domain.Stop, error)
	Departures(ctx context.Context, stopID string, n int) ([]domain.Departure, error)
}

type cached struct {
	list []domain.Departure
	at   time.Time
}

type transit struct {
	provider Provider
	clock    kernel.Clock
	mu       sync.Mutex
	cache    map[string]cached
}

// NewTransit returns the transit use cases.
func NewTransit(p Provider, clock kernel.Clock) Transit {
	return &transit{provider: p, clock: clock, cache: map[string]cached{}}
}

func unavailable() error {
	return kernel.NewPublicError(kernel.ErrUnavailable, "The departure information is not available right now.")
}

// SearchStops implements Transit.
func (uc *transit) SearchStops(ctx context.Context, query string) ([]domain.Stop, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < domain.MinQueryLength {
		return []domain.Stop{}, nil
	}
	stops, err := uc.provider.SearchStops(ctx, query)
	if err != nil {
		return nil, unavailable()
	}
	return stops, nil
}

// StopsNear implements Transit. It runs only when the person picks a place
// in the widget settings, so it is not cached.
func (uc *transit) StopsNear(ctx context.Context, lat, lon float64) ([]domain.Stop, error) {
	if !domain.ValidLocation(lat, lon) {
		return nil, domain.ErrInvalidLocation
	}
	stops, err := uc.provider.StopsNear(ctx, lat, lon)
	if err != nil {
		return nil, unavailable()
	}
	return stops, nil
}

// Departures implements Transit. Stale data is preferred over an error.
func (uc *transit) Departures(ctx context.Context, stopID string, n int) ([]domain.Departure, error) {
	if !domain.ValidStopID(stopID) {
		return nil, domain.ErrInvalidStop
	}
	now := uc.clock.Now()
	uc.mu.Lock()
	c, ok := uc.cache[stopID]
	uc.mu.Unlock()
	if !ok || now.Sub(c.at) >= cacheTTL {
		list, err := uc.provider.Departures(ctx, stopID, domain.MaxDepartures)
		if err != nil && !ok {
			return nil, unavailable()
		}
		if err == nil {
			c = cached{list: list, at: now}
			uc.mu.Lock()
			uc.cache[stopID] = c
			for id, e := range uc.cache { // forget stops no widget asked for in a while
				if now.Sub(e.at) > time.Hour {
					delete(uc.cache, id)
				}
			}
			uc.mu.Unlock()
		}
	}
	return domain.Upcoming(c.list, now, n), nil
}
