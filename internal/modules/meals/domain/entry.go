package domain

import (
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of a meal plan entry.
const (
	MaxDishLength       = 120
	MaxLinkLength       = 500
	MaxIngredients      = 40
	MaxIngredientLength = 100
	// MaxRangeDays bounds how many days one request may list.
	MaxRangeDays = 42
)

// Meal is the optional time of day of a dish. The empty meal sorts last,
// so a plan with only dinners stays readable.
type Meal string

// The meals of a day, in the order they are shown.
const (
	Breakfast Meal = "breakfast"
	Lunch     Meal = "lunch"
	Dinner    Meal = "dinner"
	AnyMeal   Meal = ""
)

// Order is the position of the meal within a day.
func (m Meal) Order() int {
	switch m {
	case Breakfast:
		return 0
	case Lunch:
		return 1
	case Dinner:
		return 2
	default:
		return 3
	}
}

// ErrNotFound is returned for unknown entries and for entries the viewer
// may not see.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The dish does not exist.")

// Entry is one dish planned for a day (#682).
type Entry struct {
	ID   string
	Day  time.Time // midnight UTC; only the date counts
	Meal Meal
	Dish string
	Link string
	// Ingredients are the shopping list lines of the dish.
	Ingredients []string
	Access      kernel.Access
	CreatedAt   time.Time
}

// Input is what a person enters for a dish.
type Input struct {
	Date        string // YYYY-MM-DD
	Meal        string
	Dish        string
	Link        string
	Ingredients []string
}

// New validates the input and creates an entry.
func New(id string, in Input, access kernel.Access, now time.Time) (Entry, error) {
	e := Entry{ID: id, Access: access, CreatedAt: now}
	if err := e.apply(in); err != nil {
		return Entry{}, err
	}
	return e, nil
}

// Update changes everything but the owner.
func (e Entry) Update(in Input) (Entry, error) {
	err := e.apply(in)
	return e, err
}

func (e *Entry) apply(in Input) error {
	day, err := time.Parse(time.DateOnly, strings.TrimSpace(in.Date))
	if err != nil {
		return kernel.NewValidationError("date", "Enter a date like 2026-12-24.")
	}
	meal := Meal(strings.TrimSpace(in.Meal))
	if meal.Order() == 3 && meal != AnyMeal {
		return kernel.NewValidationError("meal", "The meal must be breakfast, lunch, dinner or empty.")
	}
	dish := strings.TrimSpace(in.Dish)
	if dish == "" || utf8.RuneCountInString(dish) > MaxDishLength {
		return kernel.NewValidationError("dish", "The dish must be between 1 and 120 characters long.")
	}
	link, err := normaliseLink(in.Link)
	if err != nil {
		return err
	}
	ingredients, err := normaliseIngredients(in.Ingredients)
	if err != nil {
		return err
	}
	e.Day, e.Meal, e.Dish, e.Link, e.Ingredients = day, meal, dish, link, ingredients
	return nil
}

// normaliseLink accepts an empty link or an absolute http(s) URL; other
// schemes are refused because the widget renders the link as an anchor.
func normaliseLink(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(raw) > MaxLinkLength {
		return "", kernel.NewValidationError("link", "The link must be a web address starting with http:// or https://.")
	}
	return raw, nil
}

// normaliseIngredients trims the lines and drops empty ones.
func normaliseIngredients(lines []string) ([]string, error) {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if utf8.RuneCountInString(l) > MaxIngredientLength || strings.ContainsAny(l, "\r\n") {
			return nil, kernel.NewValidationError("ingredients", "An ingredient may be at most 100 characters long.")
		}
		out = append(out, l)
	}
	if len(out) > MaxIngredients {
		return nil, kernel.NewValidationError("ingredients", "A dish may have at most 40 ingredients.")
	}
	return out, nil
}

// Date is the day as YYYY-MM-DD.
func (e Entry) Date() string { return e.Day.Format(time.DateOnly) }

// Less orders entries by day, then meal, then creation.
func Less(a, b Entry) bool {
	if !a.Day.Equal(b.Day) {
		return a.Day.Before(b.Day)
	}
	if a.Meal.Order() != b.Meal.Order() {
		return a.Meal.Order() < b.Meal.Order()
	}
	return a.CreatedAt.Before(b.CreatedAt)
}

// Range is a span of whole days [From, From+Days).
type Range struct {
	From time.Time
	Days int
}

// NewRange validates a requested span; an empty from means today.
func NewRange(from string, days int, today time.Time) (Range, error) {
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if from = strings.TrimSpace(from); from != "" {
		d, err := time.Parse(time.DateOnly, from)
		if err != nil {
			return Range{}, kernel.NewValidationError("from", "Enter a date like 2026-12-24.")
		}
		start = d
	}
	if days < 1 || days > MaxRangeDays {
		return Range{}, kernel.NewValidationError("days", "Ask for 1 to 42 days.")
	}
	return Range{From: start, Days: days}, nil
}

// End is the first day after the range.
func (r Range) End() time.Time { return r.From.AddDate(0, 0, r.Days) }
