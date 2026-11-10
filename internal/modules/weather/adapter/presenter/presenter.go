package presenter

import (
	"time"

	"trisse-home/internal/modules/weather/usecase"
)

// DayDTO is one forecast day.
type DayDTO struct {
	Date                     string  `json:"date"`
	TemperatureMin           float64 `json:"temperatureMin"`
	TemperatureMax           float64 `json:"temperatureMax"`
	WeatherCode              int     `json:"weatherCode"`
	PrecipitationProbability float64 `json:"precipitationProbability"`
}

// ForecastDTO is returned by GET /api/weather.
type ForecastDTO struct {
	Lat                 float64   `json:"lat"`
	Lon                 float64   `json:"lon"`
	Timezone            string    `json:"timezone"`
	Temperature         float64   `json:"temperature"`
	ApparentTemperature float64   `json:"apparentTemperature"`
	Humidity            float64   `json:"humidity"`
	WindSpeed           float64   `json:"windSpeed"`
	WeatherCode         int       `json:"weatherCode"`
	IsDay               bool      `json:"isDay"`
	Days                []DayDTO  `json:"days"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// PlaceDTO is a geocoding result returned by GET /api/weather/places.
type PlaceDTO struct {
	Name    string  `json:"name"`
	Region  string  `json:"region"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// Forecast maps a ForecastResponse.
func Forecast(r usecase.ForecastResponse) ForecastDTO {
	dto := ForecastDTO{
		Lat: r.Lat, Lon: r.Lon, Timezone: r.Timezone, Temperature: r.Temperature,
		ApparentTemperature: r.ApparentTemperature, Humidity: r.Humidity, WindSpeed: r.WindSpeed,
		WeatherCode: r.WeatherCode, IsDay: r.IsDay, UpdatedAt: r.UpdatedAt, Days: make([]DayDTO, 0, len(r.Days)),
	}
	for _, d := range r.Days {
		dto.Days = append(dto.Days, DayDTO(d))
	}
	return dto
}

// Places maps geocoding results.
func Places(ps []usecase.PlaceView) []PlaceDTO {
	out := make([]PlaceDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, PlaceDTO(p))
	}
	return out
}
