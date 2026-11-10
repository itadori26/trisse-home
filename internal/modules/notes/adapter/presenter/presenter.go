// Package presenter maps notes responses to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/notes/usecase"
)

// NoteDTO is a note in API responses.
type NoteDTO struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Pinned    bool      `json:"pinned"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Note maps a NoteView.
func Note(v usecase.NoteView) NoteDTO { return NoteDTO(v) }

// Notes maps a list of NoteViews.
func Notes(vs []usecase.NoteView) []NoteDTO {
	out := make([]NoteDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Note(v))
	}
	return out
}
