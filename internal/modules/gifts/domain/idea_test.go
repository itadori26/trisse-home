package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestNew(t *testing.T) {
	now := time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)
	i, err := New("g1", "anna", Input{Person: " Grandma   Hilde ", Title: " Audiobook ", Link: "https://example.org/book"}, now)
	if err != nil || i.Person != "Grandma Hilde" || i.Title != "Audiobook" || i.Owner != "anna" || i.Bought {
		t.Fatalf("%+v %v", i, err)
	}
	for name, in := range map[string]Input{
		"no person":   {Title: "x"},
		"long person": {Person: strings.Repeat("x", 61), Title: "x"},
		"no title":    {Person: "Ben"},
		"script link": {Person: "Ben", Title: "x", Link: "javascript:alert(1)"},
		"long note":   {Person: "Ben", Title: "x", Note: strings.Repeat("x", 501)},
	} {
		if _, err := New("g", "anna", in, now); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	u, err := i.Update(Input{Person: "Grandma Hilde", Title: "Audiobook", Bought: true})
	if err != nil || !u.Bought || u.Owner != "anna" || u.Link != "" {
		t.Fatalf("%+v %v", u, err)
	}
}
