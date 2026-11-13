// Package presenter maps focus statistics to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/focus/domain"
)

// DayDTO is the statistic of one day.
type DayDTO struct {
	// Date is YYYY-MM-DD in the server's time zone.
	Date     string `json:"date"`
	Sessions int    `json:"sessions"`
	Minutes  int    `json:"minutes"`
}

// WeekDTO is the answer of GET /api/focus/stats.
type WeekDTO struct {
	// Days are the last seven days including today, oldest first.
	Days []DayDTO `json:"days"`
}

// Week maps the statistics.
func Week(days []domain.Day) WeekDTO {
	out := WeekDTO{Days: make([]DayDTO, 0, len(days))}
	for _, d := range days {
		out.Days = append(out.Days, DayDTO{Date: d.Date.Format(time.DateOnly), Sessions: d.Sessions, Minutes: d.Minutes})
	}
	return out
}
