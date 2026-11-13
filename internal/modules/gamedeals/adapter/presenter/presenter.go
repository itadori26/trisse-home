// Package presenter maps game deals to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"strconv"
	"time"

	"trisse-home/internal/modules/gamedeals/domain"
)

// SettingsDTO is returned by the settings endpoints.
type SettingsDTO struct {
	SteamID     string `json:"steamId"`
	Country     string `json:"country"`
	MinDiscount int    `json:"minDiscount"`
	Epic        bool   `json:"epic"`
}

// DealDTO is a discounted wishlist game.
type DealDTO struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	DiscountPct int    `json:"discountPct"`
	FinalPrice  string `json:"finalPrice"`
	// EndsAt is null when the store does not say.
	EndsAt *time.Time `json:"endsAt"`
}

// FreeGameDTO is a free game promotion.
type FreeGameDTO struct {
	Title string    `json:"title"`
	URL   string    `json:"url"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// OverviewDTO is returned by GET /api/gamedeals/overview.
type OverviewDTO struct {
	Deals    []DealDTO     `json:"deals"`
	FreeNow  []FreeGameDTO `json:"freeNow"`
	FreeNext []FreeGameDTO `json:"freeNext"`
	Errors   []string      `json:"errors"`
}

// Settings maps the settings.
func Settings(s domain.Settings) SettingsDTO { return SettingsDTO(s) }

// Overview maps the overview.
func Overview(o domain.Overview) OverviewDTO {
	dto := OverviewDTO{Deals: make([]DealDTO, 0, len(o.Deals)), FreeNow: free(o.FreeNow), FreeNext: free(o.FreeNext), Errors: o.Errors}
	if dto.Errors == nil {
		dto.Errors = []string{}
	}
	for _, d := range o.Deals {
		deal := DealDTO{Name: d.Name, URL: "https://store.steampowered.com/app/" + strconv.Itoa(d.AppID), DiscountPct: d.DiscountPct, FinalPrice: d.FinalPrice}
		if !d.EndsAt.IsZero() {
			t := d.EndsAt
			deal.EndsAt = &t
		}
		dto.Deals = append(dto.Deals, deal)
	}
	return dto
}

func free(list []domain.FreeGame) []FreeGameDTO {
	out := make([]FreeGameDTO, 0, len(list))
	for _, g := range list {
		out = append(out, FreeGameDTO(g))
	}
	return out
}
