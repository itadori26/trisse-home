package presenter

import "trisse-home/internal/modules/todos/usecase"

// TodoDTO is a to-do in API responses. Due is a date (YYYY-MM-DD) or "".
type TodoDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Due         string `json:"due"`
	Done        bool   `json:"done"`
	Overdue     bool   `json:"overdue"`
	Position    int    `json:"position"`
	CompletedAt string `json:"completedAt"`
	// Repeat is the recurrence (kind "" = none, #68).
	Repeat RepeatDTO `json:"repeat"`
	// Reminder fires at "HH:MM" on the due date or days before ("" = none).
	Reminder ReminderDTO `json:"reminder"`
	// Version counts the stored changes; send it back with an update to
	// detect concurrent changes (#80).
	Version int `json:"version"`
}

// RepeatDTO describes a recurrence.
type RepeatDTO struct {
	Kind  string `json:"kind"`
	Every int    `json:"every"`
	Days  []int  `json:"days"`
}

// ReminderDTO describes a reminder.
type ReminderDTO struct {
	At         string `json:"at"`
	DaysBefore int    `json:"daysBefore"`
}

const dateLayout = "2006-01-02"

// Todo maps a TodoView.
func Todo(v usecase.TodoView) TodoDTO {
	dto := TodoDTO{ID: v.ID, Title: v.Title, Done: v.Done, Overdue: v.Overdue, Position: v.Position,
		Repeat:   RepeatDTO{Kind: v.Repeat.Kind, Every: v.Repeat.Every, Days: append([]int{}, v.Repeat.Days...)},
		Reminder: ReminderDTO{At: v.Reminder.At, DaysBefore: v.Reminder.DaysBefore}, Version: v.Version}
	if !v.Due.IsZero() {
		dto.Due = v.Due.Format(dateLayout)
	}
	if !v.CompletedAt.IsZero() {
		dto.CompletedAt = v.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return dto
}

// Todos maps a list of TodoViews.
func Todos(vs []usecase.TodoView) []TodoDTO {
	out := make([]TodoDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Todo(v))
	}
	return out
}
