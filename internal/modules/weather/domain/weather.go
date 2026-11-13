package domain

import (
	"fmt"
	"math"
	"time"

	"trisse-home/internal/shared/kernel"
)

// coordinatePrecision rounds coordinates to ~1 km so nearby widgets share a
// cache entry and an external request.
const coordinatePrecision = 100

// Location is a validated pair of WGS84 coordinates.
type Location struct {
	lat, lon float64
}

// NewLocation validates the coordinates and rounds them to ~1 km.
func NewLocation(lat, lon float64) (Location, error) {
	if math.IsNaN(lat) || lat < -90 || lat > 90 {
		return Location{}, kernel.NewValidationError("lat", "The latitude must be between -90 and 90.")
	}
	if math.IsNaN(lon) || lon < -180 || lon > 180 {
		return Location{}, kernel.NewValidationError("lon", "The longitude must be between -180 and 180.")
	}
	round := func(v float64) float64 { return math.Round(v*coordinatePrecision) / coordinatePrecision }
	return Location{lat: round(lat), lon: round(lon)}, nil
}

// Lat returns the latitude.
func (l Location) Lat() float64 { return l.lat }

// Lon returns the longitude.
func (l Location) Lon() float64 { return l.lon }

// Key identifies the location in caches.
func (l Location) Key() string { return fmt.Sprintf("%.2f,%.2f", l.lat, l.lon) }

// Current describes the weather right now. WeatherCode follows the WMO code
// table used by Open-Meteo.
type Current struct {
	Temperature         float64
	ApparentTemperature float64
	Humidity            float64
	WindSpeed           float64
	// WindDirection is where the wind comes from in degrees (0 = north,
	// 90 = east), in [0, 360). It is only meaningful when HasWindDirection.
	WindDirection    float64
	HasWindDirection bool
	WeatherCode      int
	IsDay            bool
}

// NormalizeDegrees maps any angle into [0, 360).
func NormalizeDegrees(degrees float64) float64 {
	d := math.Mod(degrees, 360)
	if d < 0 {
		d += 360
	}
	return d
}

// Day is the forecast for one day.
type Day struct {
	Date                     time.Time
	TemperatureMin           float64
	TemperatureMax           float64
	WeatherCode              int
	PrecipitationProbability float64
}

// Forecast combines current conditions and the daily outlook.
type Forecast struct {
	Location  Location
	Timezone  string
	Current   Current
	Days      []Day
	FetchedAt time.Time
}

// Place is a geocoding search result.
type Place struct {
	Name    string
	Region  string
	Country string
	Lat     float64
	Lon     float64
}
