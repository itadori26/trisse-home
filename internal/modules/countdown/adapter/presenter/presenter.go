// Package presenter maps countdowns to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/countdown/domain"
)

// CountdownDTO is one countdown in API responses.
type CountdownDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Emoji string `json:"emoji"`
	// Date is YYYY-MM-DD.
	Date string `json:"date"`
	// Time is HH:MM, or empty for the whole day.
	Time      string    `json:"time"`
	CreatedAt time.Time `json:"createdAt"`
}

// Countdown maps one countdown.
func Countdown(c domain.Countdown) CountdownDTO {
	return CountdownDTO{ID: c.ID, Title: c.Title, Emoji: c.Emoji, Date: c.Date(), Time: c.Time(), CreatedAt: c.CreatedAt}
}

// Countdowns maps a list of countdowns.
func Countdowns(list []domain.Countdown) []CountdownDTO {
	out := make([]CountdownDTO, 0, len(list))
	for _, c := range list {
		out = append(out, Countdown(c))
	}
	return out
}
