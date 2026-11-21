// Package presenter maps chores to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/chores/usecase"
)

// ChoreDTO is one chore in API responses.
type ChoreDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	EveryDays int    `json:"everyDays"`
	// Due is the next due day, YYYY-MM-DD.
	Due string `json:"due"`
	// Assignees take turns in this order; Assignee is whose turn it is.
	Assignees []string `json:"assignees"`
	Assignee  string   `json:"assignee"`
	// DueNow is true when the chore is due today or overdue.
	DueNow  bool `json:"dueNow"`
	Overdue bool `json:"overdue"`
	// LastDoneAt is omitted while the chore was never done.
	LastDoneAt *time.Time `json:"lastDoneAt,omitempty"`
	LastDoneBy string     `json:"lastDoneBy"`
	// Visibility is "private", "everyone" or "group:<id>" (ADR 0018).
	Visibility string `json:"visibility"`
	// Mine is true when the viewer created the chore and may change it.
	Mine bool `json:"mine"`
}

// Chore maps one chore.
func Chore(v usecase.ChoreView) ChoreDTO {
	assignees := v.Assignees
	if assignees == nil {
		assignees = []string{}
	}
	dto := ChoreDTO{ID: v.ID, Title: v.Title, EveryDays: v.EveryDays, Due: v.DueDate(), Assignees: assignees, Assignee: v.Assignee(),
		DueNow: v.DueNow, Overdue: v.Overdue, LastDoneBy: v.LastDoneBy, Visibility: string(v.Access.Visibility), Mine: v.Mine}
	if !v.LastDoneAt.IsZero() {
		at := v.LastDoneAt
		dto.LastDoneAt = &at
	}
	return dto
}

// Chores maps a list of chores.
func Chores(list []usecase.ChoreView) []ChoreDTO {
	out := make([]ChoreDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Chore(v))
	}
	return out
}
