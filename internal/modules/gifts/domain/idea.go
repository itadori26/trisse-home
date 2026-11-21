package domain

import (
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of a gift idea.
const (
	MaxPersonLength = 60
	MaxTitleLength  = 120
	MaxLinkLength   = 500
	MaxNoteLength   = 500
)

// ErrNotFound is returned for unknown ideas and for ideas of someone else;
// the two are not told apart, so an idea's existence stays private.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The gift idea does not exist.")

// Idea is a gift idea for a person (#684). Ideas are always private to
// their author: there is no visibility to choose.
type Idea struct {
	ID     string
	Person string
	Title  string
	Link   string
	Note   string
	Bought bool
	// Owner is the author, the only account that ever sees the idea.
	Owner     string
	CreatedAt time.Time
}

// Input is what a person enters for an idea.
type Input struct {
	Person string
	Title  string
	Link   string
	Note   string
	Bought bool
}

// New validates the input and creates an idea of owner.
func New(id, owner string, in Input, now time.Time) (Idea, error) {
	i := Idea{ID: id, Owner: owner, CreatedAt: now}
	if err := i.apply(in); err != nil {
		return Idea{}, err
	}
	return i, nil
}

// Update changes everything but the author.
func (i Idea) Update(in Input) (Idea, error) {
	err := i.apply(in)
	return i, err
}

func (i *Idea) apply(in Input) error {
	person := strings.Join(strings.Fields(in.Person), " ")
	if person == "" || utf8.RuneCountInString(person) > MaxPersonLength {
		return kernel.NewValidationError("person", "The name must be between 1 and 60 characters long.")
	}
	title := strings.TrimSpace(in.Title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLength {
		return kernel.NewValidationError("title", "The idea must be between 1 and 120 characters long.")
	}
	link := strings.TrimSpace(in.Link)
	if link != "" {
		u, err := url.Parse(link)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(link) > MaxLinkLength {
			return kernel.NewValidationError("link", "The link must be a web address starting with http:// or https://.")
		}
	}
	note := strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(note) > MaxNoteLength {
		return kernel.NewValidationError("note", "The note may be at most 500 characters long.")
	}
	i.Person, i.Title, i.Link, i.Note, i.Bought = person, title, link, note, in.Bought
	return nil
}
