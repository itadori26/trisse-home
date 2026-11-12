// Package presenter maps habits to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/habits/usecase"
)

// HeatCellDTO is one day of the 4-week heatmap.
type HeatCellDTO struct {
	// Date is YYYY-MM-DD.
	Date string `json:"date"`
	Due  bool   `json:"due"`
	Done bool   `json:"done"`
}

// HabitDTO is a habit with its statistics.
type HabitDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	// Days are the weekdays 0 (Sunday) to 6; empty = every day.
	Days      []int         `json:"days"`
	DoneToday bool          `json:"doneToday"`
	Current   int           `json:"current"`
	Best      int           `json:"best"`
	Heatmap   []HeatCellDTO `json:"heatmap"`
}

// Habit maps one habit.
func Habit(v usecase.HabitView) HabitDTO {
	cells := make([]HeatCellDTO, 0, len(v.Stats.Heatmap))
	for _, c := range v.Stats.Heatmap {
		cells = append(cells, HeatCellDTO{Date: c.Day.Format(time.DateOnly), Due: c.Due, Done: c.Done})
	}
	return HabitDTO{ID: v.Habit.ID, Name: v.Habit.Name, Emoji: v.Habit.Emoji, Days: v.Habit.Schedule.Days(),
		DoneToday: v.Stats.DoneToday, Current: v.Stats.Current, Best: v.Stats.Best, Heatmap: cells}
}

// Habits maps a list of habits.
func Habits(list []usecase.HabitView) []HabitDTO {
	out := make([]HabitDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Habit(v))
	}
	return out
}
