package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/modules/weather/domain"
)

// Default DWD open data files of the health forecasts (#56).
const (
	DefaultPollenURL = "https://opendata.dwd.de/climate_environment/health/alerts/s31fg.json"
	DefaultUVURL     = "https://opendata.dwd.de/climate_environment/health/alerts/uvi.json"
)

// Health implements usecase.HealthProvider with the DWD open data files.
type Health struct {
	dwd        *DWD
	pollen, uv string
	berlin     *time.Location
}

// NewHealth returns the provider; it reuses the HTTP handling of the DWD client.
func NewHealth(dwd *DWD, pollenURL, uvURL string) *Health {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		berlin = time.UTC
	}
	return &Health{dwd: dwd, pollen: pollenURL, uv: uvURL, berlin: berlin}
}

type pollenFeed struct {
	LastUpdate string `json:"last_update"`
	Content    []struct {
		RegionID       int    `json:"region_id"`
		RegionName     string `json:"region_name"`
		PartregionID   int    `json:"partregion_id"`
		PartregionName string `json:"partregion_name"`
		Pollen         map[string]struct {
			Today    string `json:"today"`
			Tomorrow string `json:"tomorrow"`
		} `json:"Pollen"`
	} `json:"content"`
}

// Pollen implements usecase.HealthProvider.
func (h *Health) Pollen(ctx context.Context) ([]domain.PollenForecast, error) {
	body, _, err := h.dwd.get(ctx, h.pollen)
	if err != nil {
		return nil, err
	}
	var feed pollenFeed
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("decode pollen: %w", err)
	}
	// "2026-11-13 11:00 Uhr", local time in Germany.
	updated, _ := time.ParseInLocation("2006-01-02 15:04", strings.TrimSuffix(strings.TrimSpace(feed.LastUpdate), " Uhr"), h.berlin)
	out := make([]domain.PollenForecast, 0, len(feed.Content))
	for _, c := range feed.Content {
		region := domain.PollenRegion{ID: c.RegionID, Name: c.RegionName}
		if c.PartregionID > 0 {
			region = domain.PollenRegion{ID: c.PartregionID, Name: c.RegionName + " – " + c.PartregionName}
		}
		f := domain.PollenForecast{Region: region, UpdatedAt: updated}
		for name, v := range c.Pollen {
			t, ok := domain.ParsePollenType(name)
			if !ok {
				continue
			}
			f.Days = append(f.Days, domain.PollenDay{Type: t, Today: domain.ParsePollenLevel(v.Today), Tomorrow: domain.ParsePollenLevel(v.Tomorrow)})
		}
		f.Days = f.Of(nil) // season order
		out = append(out, f)
	}
	return out, nil
}

type uvFeed struct {
	LastUpdate string `json:"last_update"`
	Content    []struct {
		City     string `json:"city"`
		Forecast struct {
			Today    *int `json:"today"`
			Tomorrow *int `json:"tomorrow"`
		} `json:"forecast"`
	} `json:"content"`
}

// UV implements usecase.HealthProvider.
func (h *Health) UV(ctx context.Context) ([]domain.UVForecast, error) {
	body, _, err := h.dwd.get(ctx, h.uv)
	if err != nil {
		return nil, err
	}
	var feed uvFeed
	if err := json.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("decode UV index: %w", err)
	}
	// "2026-11-13T07:30:00", local time in Germany.
	updated, _ := time.ParseInLocation("2006-01-02T15:04:05", strings.TrimSpace(feed.LastUpdate), h.berlin)
	value := func(p *int) int {
		if p == nil {
			return -1
		}
		return *p
	}
	out := make([]domain.UVForecast, 0, len(feed.Content))
	for _, c := range feed.Content {
		out = append(out, domain.UVForecast{City: c.City, Today: value(c.Forecast.Today), Tomorrow: value(c.Forecast.Tomorrow), UpdatedAt: updated})
	}
	return out, nil
}
