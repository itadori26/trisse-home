// Package presenter maps gift ideas to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/gifts/domain"
)

// IdeaDTO is one gift idea in API responses. Ideas are always private, so
// there is no visibility or owner field.
type IdeaDTO struct {
	ID        string    `json:"id"`
	Person    string    `json:"person"`
	Title     string    `json:"title"`
	Link      string    `json:"link"`
	Note      string    `json:"note"`
	Bought    bool      `json:"bought"`
	CreatedAt time.Time `json:"createdAt"`
}

// Idea maps one idea.
func Idea(i domain.Idea) IdeaDTO {
	return IdeaDTO{ID: i.ID, Person: i.Person, Title: i.Title, Link: i.Link, Note: i.Note, Bought: i.Bought, CreatedAt: i.CreatedAt}
}

// Ideas maps a list of ideas.
func Ideas(list []domain.Idea) []IdeaDTO {
	out := make([]IdeaDTO, 0, len(list))
	for _, i := range list {
		out = append(out, Idea(i))
	}
	return out
}
