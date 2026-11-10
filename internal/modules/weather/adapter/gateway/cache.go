// Package gateway implements the weather output ports: an Open-Meteo client
// and an in-memory forecast cache.
//
// Ring: Interface Adapters.
package gateway

import (
	"sync"
	"time"

	"trisse-home/internal/modules/weather/domain"
)

// MemoryCache is a concurrency-safe in-memory usecase.ForecastCache.
type MemoryCache struct {
	mu        sync.RWMutex
	forecasts map[string]domain.Forecast
	requested map[string]requestedLocation
}

type requestedLocation struct {
	loc domain.Location
	at  time.Time
}

// NewMemoryCache returns an empty cache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{forecasts: map[string]domain.Forecast{}, requested: map[string]requestedLocation{}}
}

// Get returns the cached forecast for key.
func (c *MemoryCache) Get(key string) (domain.Forecast, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	f, ok := c.forecasts[key]
	return f, ok
}

// Put stores a forecast.
func (c *MemoryCache) Put(f domain.Forecast) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.forecasts[f.Location.Key()] = f
}

// Touch records a request for loc.
func (c *MemoryCache) Touch(loc domain.Location, at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requested[loc.Key()] = requestedLocation{loc: loc, at: at}
}

// Locations returns locations requested since the given time and evicts the rest.
func (c *MemoryCache) Locations(since time.Time) []domain.Location {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]domain.Location, 0, len(c.requested))
	for key, r := range c.requested {
		if r.at.Before(since) {
			delete(c.requested, key)
			delete(c.forecasts, key)
			continue
		}
		out = append(out, r.loc)
	}
	return out
}
