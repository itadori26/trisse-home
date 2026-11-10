package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/shared/kernel"
)

// Default Open-Meteo endpoints.
const (
	DefaultForecastURL  = "https://api.open-meteo.com/v1/forecast"
	DefaultGeocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastDays        = 5
	maxPlaces           = 6
	maxResponseBytes    = 1 << 20
)

// OpenMeteo implements usecase.ForecastProvider against the Open-Meteo API.
type OpenMeteo struct {
	client       *http.Client
	forecastURL  string
	geocodingURL string
	language     string
}

// NewOpenMeteo returns a client. Base URLs are injectable for tests.
func NewOpenMeteo(client *http.Client, forecastURL, geocodingURL, language string) *OpenMeteo {
	return &OpenMeteo{client: client, forecastURL: forecastURL, geocodingURL: geocodingURL, language: language}
}

// forecastPayload is the subset of the Open-Meteo forecast response we use.
type forecastPayload struct {
	Timezone string `json:"timezone"`
	Current  struct {
		Temperature         float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		Humidity            float64 `json:"relative_humidity_2m"`
		WindSpeed           float64 `json:"wind_speed_10m"`
		WeatherCode         int     `json:"weather_code"`
		IsDay               int     `json:"is_day"`
	} `json:"current"`
	Daily struct {
		Time           []string   `json:"time"`
		WeatherCode    []int      `json:"weather_code"`
		TemperatureMax []float64  `json:"temperature_2m_max"`
		TemperatureMin []float64  `json:"temperature_2m_min"`
		Precipitation  []*float64 `json:"precipitation_probability_max"`
	} `json:"daily"`
}

// Forecast fetches current conditions and a 5-day outlook.
func (o *OpenMeteo) Forecast(ctx context.Context, loc domain.Location) (domain.Forecast, error) {
	q := url.Values{}
	q.Set("latitude", strconv.FormatFloat(loc.Lat(), 'f', 2, 64))
	q.Set("longitude", strconv.FormatFloat(loc.Lon(), 'f', 2, 64))
	q.Set("current", "temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,weather_code,is_day")
	q.Set("daily", "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max")
	q.Set("timezone", "auto")
	q.Set("forecast_days", strconv.Itoa(forecastDays))
	var p forecastPayload
	if err := o.getJSON(ctx, o.forecastURL+"?"+q.Encode(), &p); err != nil {
		return domain.Forecast{}, err
	}
	f := domain.Forecast{
		Location: loc, Timezone: p.Timezone,
		Current: domain.Current{
			Temperature: p.Current.Temperature, ApparentTemperature: p.Current.ApparentTemperature,
			Humidity: p.Current.Humidity, WindSpeed: p.Current.WindSpeed,
			WeatherCode: p.Current.WeatherCode, IsDay: p.Current.IsDay == 1,
		},
	}
	d := p.Daily
	for i, day := range d.Time {
		if i >= len(d.WeatherCode) || i >= len(d.TemperatureMax) || i >= len(d.TemperatureMin) {
			break
		}
		date, err := time.Parse(time.DateOnly, day)
		if err != nil {
			return domain.Forecast{}, fmt.Errorf("parse forecast date %q: %w", day, err)
		}
		var precip float64
		if i < len(d.Precipitation) && d.Precipitation[i] != nil {
			precip = *d.Precipitation[i]
		}
		f.Days = append(f.Days, domain.Day{
			Date: date, WeatherCode: d.WeatherCode[i], TemperatureMax: d.TemperatureMax[i],
			TemperatureMin: d.TemperatureMin[i], PrecipitationProbability: precip,
		})
	}
	return f, nil
}

type geocodingPayload struct {
	Results []struct {
		Name      string  `json:"name"`
		Admin1    string  `json:"admin1"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

// SearchPlaces looks up places by name.
func (o *OpenMeteo) SearchPlaces(ctx context.Context, query string) ([]domain.Place, error) {
	q := url.Values{}
	q.Set("name", query)
	q.Set("count", strconv.Itoa(maxPlaces))
	q.Set("language", o.language)
	q.Set("format", "json")
	var p geocodingPayload
	if err := o.getJSON(ctx, o.geocodingURL+"?"+q.Encode(), &p); err != nil {
		return nil, err
	}
	places := make([]domain.Place, 0, len(p.Results))
	for _, r := range p.Results {
		places = append(places, domain.Place{Name: r.Name, Region: r.Admin1, Country: r.Country, Lat: r.Latitude, Lon: r.Longitude})
	}
	return places, nil
}

func (o *OpenMeteo) getJSON(ctx context.Context, rawURL string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("open-meteo request: %w: %w", kernel.ErrUnavailable, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("open-meteo returned %s: %w", res.Status, kernel.ErrUnavailable)
	}
	if err := json.NewDecoder(http.MaxBytesReader(nil, res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("decode open-meteo response: %w", err)
	}
	return nil
}
