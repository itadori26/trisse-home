// Package presenter maps fuel use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"

	"trisse-home/internal/modules/fuel/domain"
	"trisse-home/internal/modules/fuel/usecase"
)

// FavouriteDTO is a favourite station.
type FavouriteDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SettingsDTO is returned by the settings endpoints; the API key is never returned.
type SettingsDTO struct {
	Configured bool           `json:"configured"`
	HasAPIKey  bool           `json:"hasApiKey"`
	Lat        float64        `json:"lat"`
	Lon        float64        `json:"lon"`
	Place      string         `json:"place"`
	RadiusKm   float64        `json:"radiusKm"`
	Fuel       string         `json:"fuel" tstype:"'e5' | 'e10' | 'diesel'"`
	Favourites []FavouriteDTO `json:"favourites"`
}

// StationDTO is a station with the price of the chosen fuel.
type StationDTO struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Brand      string  `json:"brand"`
	Address    string  `json:"address"`
	DistanceKm float64 `json:"distanceKm"`
	// Price is null when the station reports none.
	Price *float64 `json:"price"`
	Open  bool     `json:"open"`
}

// OverviewDTO is returned by GET /api/fuel/overview.
type OverviewDTO struct {
	Fuel       string       `json:"fuel" tstype:"'e5' | 'e10' | 'diesel'"`
	Place      string       `json:"place"`
	Cheapest   []StationDTO `json:"cheapest"`
	Favourites []StationDTO `json:"favourites"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO {
	dto := SettingsDTO{Configured: v.Configured, HasAPIKey: v.HasAPIKey, Lat: v.Lat, Lon: v.Lon, Place: v.Place, RadiusKm: v.RadiusKm,
		Fuel: v.Fuel, Favourites: make([]FavouriteDTO, 0, len(v.Favourites))}
	for _, f := range v.Favourites {
		dto.Favourites = append(dto.Favourites, FavouriteDTO(f))
	}
	return dto
}

// Overview maps the overview.
func Overview(o domain.Overview) OverviewDTO {
	return OverviewDTO{Fuel: string(o.Fuel), Place: o.Place, Cheapest: stations(o.Cheapest), Favourites: stations(o.Favourites)}
}

func stations(list []domain.Station) []StationDTO {
	out := make([]StationDTO, 0, len(list))
	for _, s := range list {
		dto := StationDTO{ID: s.ID, Name: s.Name, Brand: s.Brand, Address: s.Address, DistanceKm: math.Round(s.DistanceKm*10) / 10, Open: s.Open}
		if s.Price > 0 {
			p := s.Price
			dto.Price = &p
		}
		out = append(out, dto)
	}
	return out
}
