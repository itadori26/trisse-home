package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of a chore.
const (
	MaxTitleLength    = 80
	MaxEveryDays      = 365
	MaxAssignees      = 8
	MaxAssigneeLength = 40
)

// ErrNotFound is returned for unknown chores and for chores the viewer may
// not see.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The chore does not exist.")

// Chore is a recurring household task (#683). Its assignees take turns: each
// completion passes it to the next name in the list.
type Chore struct {
	ID    string
	Title string
	// EveryDays is the repeat interval: 1 = daily, 7 = weekly.
	EveryDays int
	// Due is the next day the chore is due (midnight UTC; only the date
	// counts).
	Due       time.Time
	Assignees []string
	// Turn is the index of the assignee whose turn it is.
	Turn int
	// LastDoneAt and LastDoneBy describe the last completion; LastDoneBy is
	// the assignee whose turn it was, empty without assignees.
	LastDoneAt time.Time
	LastDoneBy string
	Access     kernel.Access
	CreatedAt  time.Time
}

// Input is what a person enters for a chore.
type Input struct {
	Title     string
	EveryDays int
	Due       string // YYYY-MM-DD
	Assignees []string
}

// New validates the input and creates a chore; the first assignee starts.
func New(id string, in Input, access kernel.Access, now time.Time) (Chore, error) {
	c := Chore{ID: id, Access: access, CreatedAt: now}
	if err := c.apply(in); err != nil {
		return Chore{}, err
	}
	return c, nil
}

// Update changes title, interval, due day and assignees. The turn stays
// with the same position, wrapped when the list got shorter.
func (c Chore) Update(in Input) (Chore, error) {
	if err := c.apply(in); err != nil {
		return Chore{}, err
	}
	if len(c.Assignees) == 0 {
		c.Turn = 0
	} else {
		c.Turn %= len(c.Assignees)
	}
	return c, nil
}

func (c *Chore) apply(in Input) error {
	title := strings.TrimSpace(in.Title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLength {
		return kernel.NewValidationError("title", "The chore must be between 1 and 80 characters long.")
	}
	if in.EveryDays < 1 || in.EveryDays > MaxEveryDays {
		return kernel.NewValidationError("everyDays", "Repeat a chore every 1 to 365 days.")
	}
	due, err := time.Parse(time.DateOnly, strings.TrimSpace(in.Due))
	if err != nil {
		return kernel.NewValidationError("due", "Enter a date like 2026-12-24.")
	}
	assignees := make([]string, 0, len(in.Assignees))
	for _, a := range in.Assignees {
		a = strings.TrimSpace(a)
		if a == "" {
			continue
		}
		if utf8.RuneCountInString(a) > MaxAssigneeLength {
			return kernel.NewValidationError("assignees", "A name may be at most 40 characters long.")
		}
		assignees = append(assignees, a)
	}
	if len(assignees) > MaxAssignees {
		return kernel.NewValidationError("assignees", "A chore can rotate between at most 8 people.")
	}
	c.Title, c.EveryDays, c.Due, c.Assignees = title, in.EveryDays, due, assignees
	return nil
}

// Assignee is whose turn it is, or empty without assignees.
func (c Chore) Assignee() string {
	if len(c.Assignees) == 0 {
		return ""
	}
	return c.Assignees[c.Turn%len(c.Assignees)]
}

// DueOn reports whether the chore is due on day or earlier.
func (c Chore) DueOn(day time.Time) bool { return !c.Due.After(dateOf(day)) }

// OverdueOn reports whether the chore was due before day.
func (c Chore) OverdueOn(day time.Time) bool { return c.Due.Before(dateOf(day)) }

// Done records a completion on day at the moment now: the chore moves on by
// its interval and passes to the next assignee. A chore that was overdue
// counts its next round from day, so it does not come back at once.
func (c Chore) Done(day, now time.Time) Chore {
	today := dateOf(day)
	c.LastDoneAt, c.LastDoneBy = now, c.Assignee()
	next := c.Due.AddDate(0, 0, c.EveryDays)
	if !next.After(today) {
		next = today.AddDate(0, 0, c.EveryDays)
	}
	c.Due = next
	if len(c.Assignees) > 0 {
		c.Turn = (c.Turn + 1) % len(c.Assignees)
	}
	return c
}

// DueDate is the due day as YYYY-MM-DD.
func (c Chore) DueDate() string { return c.Due.Format(time.DateOnly) }

// dateOf is the calendar day of t in its own location, as midnight UTC like
// the stored due days.
func dateOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
