package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestNew(t *testing.T) {
	now := time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)
	ok := Input{Date: "2026-11-23", Meal: "dinner", Dish: " Lasagne ", Link: "https://example.org/lasagne", Ingredients: []string{" pasta ", "", "tomatoes"}}
	e, err := New("e1", ok, kernel.Access{Owner: "anna"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if e.Dish != "Lasagne" || e.Date() != "2026-11-23" || e.Meal != Dinner || !slices.Equal(e.Ingredients, []string{"pasta", "tomatoes"}) {
		t.Fatalf("got %+v", e)
	}
	for name, in := range map[string]Input{
		"no date":         {Dish: "Soup"},
		"bad meal":        {Date: "2026-11-23", Meal: "brunch", Dish: "Soup"},
		"no dish":         {Date: "2026-11-23", Dish: " "},
		"long dish":       {Date: "2026-11-23", Dish: strings.Repeat("x", 121)},
		"script link":     {Date: "2026-11-23", Dish: "Soup", Link: "javascript:alert(1)"},
		"relative link":   {Date: "2026-11-23", Dish: "Soup", Link: "/recipes/1"},
		"long ingredient": {Date: "2026-11-23", Dish: "Soup", Ingredients: []string{strings.Repeat("x", 101)}},
		"many":            {Date: "2026-11-23", Dish: "Soup", Ingredients: slices.Repeat([]string{"salt"}, 41)},
	} {
		if _, err := New("e", in, kernel.Access{}, now); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestLess(t *testing.T) {
	day := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	at := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	list := []Entry{
		{ID: "any", Day: day, CreatedAt: at},
		{ID: "next", Day: day.AddDate(0, 0, 1), Meal: Breakfast},
		{ID: "dinner", Day: day, Meal: Dinner},
		{ID: "lunch", Day: day, Meal: Lunch},
		{ID: "any-earlier", Day: day, CreatedAt: at.Add(-time.Hour)},
	}
	slices.SortFunc(list, func(a, b Entry) int {
		if Less(a, b) {
			return -1
		}
		if Less(b, a) {
			return 1
		}
		return 0
	})
	var ids []string
	for _, e := range list {
		ids = append(ids, e.ID)
	}
	if want := []string{"lunch", "dinner", "any-earlier", "any", "next"}; !slices.Equal(ids, want) {
		t.Fatalf("order %v, want %v", ids, want)
	}
}

func TestNewRange(t *testing.T) {
	today := time.Date(2026, 11, 21, 23, 30, 0, 0, time.FixedZone("CET", 3600))
	r, err := NewRange("", 7, today)
	if err != nil || r.From.Format(time.DateOnly) != "2026-11-21" || r.End().Format(time.DateOnly) != "2026-11-28" {
		t.Fatalf("%+v %v", r, err)
	}
	if _, err := NewRange("2026-11-23", 43, today); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("too many days accepted")
	}
	if _, err := NewRange("23.11.2026", 7, today); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("bad date accepted")
	}
}
