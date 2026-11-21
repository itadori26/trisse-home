// Package presenter maps meal plan entries to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/meals/usecase"
)

// EntryDTO is one planned dish in API responses.
type EntryDTO struct {
	ID string `json:"id"`
	// Date is YYYY-MM-DD.
	Date string `json:"date"`
	// Meal is "breakfast", "lunch", "dinner" or empty.
	Meal        string   `json:"meal"`
	Dish        string   `json:"dish"`
	Link        string   `json:"link"`
	Ingredients []string `json:"ingredients"`
	// Visibility is "private", "everyone" or "group:<id>" (ADR 0018).
	Visibility string `json:"visibility"`
	// Mine is true when the viewer created the dish and may change it.
	Mine      bool      `json:"mine"`
	CreatedAt time.Time `json:"createdAt"`
}

// Entry maps one dish.
func Entry(v usecase.EntryView) EntryDTO {
	ingredients := v.Ingredients
	if ingredients == nil {
		ingredients = []string{}
	}
	return EntryDTO{ID: v.ID, Date: v.Date(), Meal: string(v.Meal), Dish: v.Dish, Link: v.Link, Ingredients: ingredients,
		Visibility: string(v.Access.Visibility), Mine: v.Mine, CreatedAt: v.CreatedAt}
}

// Entries maps a list of dishes.
func Entries(list []usecase.EntryView) []EntryDTO {
	out := make([]EntryDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Entry(v))
	}
	return out
}
