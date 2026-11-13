package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/shared/kernel"
)

// Cache timing rules.
const (
	// MaxForecastAge is how old a cached forecast may be before it is
	// fetched again on request.
	MaxForecastAge = 30 * time.Minute
	// UnusedLocationTTL removes locations that no widget has requested for a
	// while from the background refresh.
	UnusedLocationTTL = 24 * time.Hour
	minPlaceQuery     = 2
	maxPlaceQuery     = 100
)

// ForecastRequest is the input of GetForecast.
type ForecastRequest struct {
	Lat, Lon float64
}

// DayView is one day of the forecast.
type DayView struct {
	Date                     string
	TemperatureMin           float64
	TemperatureMax           float64
	WeatherCode              int
	PrecipitationProbability float64
}

// ForecastResponse is the output of GetForecast.
type ForecastResponse struct {
	Lat, Lon            float64
	Timezone            string
	Temperature         float64
	ApparentTemperature float64
	Humidity            float64
	WindSpeed           float64
	// WindDirection in degrees the wind comes from; HasWindDirection tells whether it is known.
	WindDirection    float64
	HasWindDirection bool
	WeatherCode      int
	IsDay            bool
	Days             []DayView
	UpdatedAt        time.Time
}

// PlaceView is a geocoding result.
type PlaceView struct {
	Name, Region, Country string
	Lat, Lon              float64
}

// Weather is the input port of the weather module.
type Weather interface {
	GetForecast(ctx context.Context, req ForecastRequest) (ForecastResponse, error)
	RefreshAll(ctx context.Context) error
	SearchPlaces(ctx context.Context, query string) ([]PlaceView, error)
}

// ForecastProvider fetches forecasts from an external service (output port).
type ForecastProvider interface {
	Forecast(ctx context.Context, loc domain.Location) (domain.Forecast, error)
	SearchPlaces(ctx context.Context, query string) ([]domain.Place, error)
}

// ForecastCache stores forecasts per location (output port).
type ForecastCache interface {
	Get(key string) (domain.Forecast, bool)
	Put(f domain.Forecast)
	// Touch records that a location was requested.
	Touch(loc domain.Location, at time.Time)
	// Locations returns the locations requested since the given time and
	// forgets all others.
	Locations(since time.Time) []domain.Location
}

type weather struct {
	provider ForecastProvider
	cache    ForecastCache
	clock    kernel.Clock
}

// NewWeather returns the weather use cases.
func NewWeather(provider ForecastProvider, cache ForecastCache, clock kernel.Clock) Weather {
	return &weather{provider: provider, cache: cache, clock: clock}
}

// GetForecast implements Weather (cache-aside).
func (uc *weather) GetForecast(ctx context.Context, req ForecastRequest) (ForecastResponse, error) {
	loc, err := domain.NewLocation(req.Lat, req.Lon)
	if err != nil {
		return ForecastResponse{}, err
	}
	now := uc.clock.Now()
	uc.cache.Touch(loc, now)
	if f, ok := uc.cache.Get(loc.Key()); ok && now.Sub(f.FetchedAt) < MaxForecastAge {
		return toResponse(f), nil
	}
	f, err := uc.fetch(ctx, loc)
	if err != nil {
		if cached, ok := uc.cache.Get(loc.Key()); ok {
			return toResponse(cached), nil // stale data beats an error
		}
		return ForecastResponse{}, err
	}
	return toResponse(f), nil
}

// RefreshAll implements Weather; it is run by the background job.
func (uc *weather) RefreshAll(ctx context.Context) error {
	var failures []string
	for _, loc := range uc.cache.Locations(uc.clock.Now().Add(-UnusedLocationTTL)) {
		if _, err := uc.fetch(ctx, loc); err != nil {
			failures = append(failures, loc.Key())
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("refresh weather for %s: %w", strings.Join(failures, "; "), kernel.ErrUnavailable)
	}
	return nil
}

func (uc *weather) fetch(ctx context.Context, loc domain.Location) (domain.Forecast, error) {
	f, err := uc.provider.Forecast(ctx, loc)
	if err != nil {
		return domain.Forecast{}, fmt.Errorf("fetch forecast: %w", err)
	}
	f.Location, f.FetchedAt = loc, uc.clock.Now()
	uc.cache.Put(f)
	return f, nil
}

// SearchPlaces implements Weather.
func (uc *weather) SearchPlaces(ctx context.Context, query string) ([]PlaceView, error) {
	query = strings.TrimSpace(query)
	if len(query) < minPlaceQuery || len(query) > maxPlaceQuery {
		return nil, kernel.NewValidationError("q", "Please enter between 2 and 100 characters.")
	}
	places, err := uc.provider.SearchPlaces(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search places: %w", err)
	}
	out := make([]PlaceView, 0, len(places))
	for _, p := range places {
		out = append(out, PlaceView{Name: p.Name, Region: p.Region, Country: p.Country, Lat: p.Lat, Lon: p.Lon})
	}
	return out, nil
}

func toResponse(f domain.Forecast) ForecastResponse {
	res := ForecastResponse{
		Lat: f.Location.Lat(), Lon: f.Location.Lon(), Timezone: f.Timezone,
		Temperature: f.Current.Temperature, ApparentTemperature: f.Current.ApparentTemperature,
		Humidity: f.Current.Humidity, WindSpeed: f.Current.WindSpeed, WeatherCode: f.Current.WeatherCode,
		WindDirection: f.Current.WindDirection, HasWindDirection: f.Current.HasWindDirection,
		IsDay: f.Current.IsDay, UpdatedAt: f.FetchedAt, Days: make([]DayView, 0, len(f.Days)),
	}
	for _, d := range f.Days {
		res.Days = append(res.Days, DayView{
			Date: d.Date.Format(time.DateOnly), TemperatureMin: d.TemperatureMin, TemperatureMax: d.TemperatureMax,
			WeatherCode: d.WeatherCode, PrecipitationProbability: d.PrecipitationProbability,
		})
	}
	return res
}
