package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/services/domain"
)

// SeedSampleData creates example service tiles if there are none (input
// port). It reports whether tiles were created so the caller can check them.
type SeedSampleData interface {
	Execute(ctx context.Context) (bool, error)
}

type seedSampleData struct{ d Deps }

// NewSeedSampleData returns the SeedSampleData use case.
func NewSeedSampleData(d Deps) SeedSampleData { return &seedSampleData{d: d} }

// sampleServices are public sites, so their status is meaningful anywhere.
var sampleServices = []ServiceInput{
	{Name: "DuckDuckGo", URL: "https://duckduckgo.com/", Icon: "🦆", Description: "Example tile – replace with your services"},
	{Name: "OpenStreetMap", URL: "https://www.openstreetmap.org/", Icon: "🗺️", Description: "Example tile"},
	{Name: "Open-Meteo", URL: "https://open-meteo.com/", Icon: "🌦️", Description: "Weather data source"},
}

// Execute implements SeedSampleData. Tiles are stored without the immediate
// health check of Create; the caller decides when to check them.
func (uc *seedSampleData) Execute(ctx context.Context) (bool, error) {
	existing, err := uc.d.Repo.List(ctx)
	if err != nil {
		return false, fmt.Errorf("list services: %w", err)
	}
	if len(existing) > 0 {
		return false, nil
	}
	for i, in := range sampleServices {
		s, err := domain.NewService(uc.d.IDs.NewID(), in.Name, in.URL, in.CheckURL, in.Icon, in.Description, in.SkipTLSVerify, i)
		if err != nil {
			return false, err
		}
		if err := uc.d.Repo.Create(ctx, s); err != nil {
			return false, fmt.Errorf("create service: %w", err)
		}
	}
	return true, nil
}
