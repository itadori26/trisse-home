package domain

import (
	"cmp"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits.
const (
	maxSummaryLength = 200
	// MaxCompletedShown is how many completed items are listed.
	MaxCompletedShown = 10
)

var itemIDPattern = regexp.MustCompile(`^[A-Za-z0-9._@%+~-]{1,200}\.ics$`)

// Errors.
var (
	ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "The shopping list is not configured yet.")
	ErrItemNotFound  = kernel.NewPublicError(kernel.ErrNotFound, "The item does not exist (any more).")
	ErrConflict      = kernel.NewPublicError(kernel.ErrConflict, "The item was changed in another app at the same time. Please try again.")
)

// Settings point to a CalDAV task list (a calendar collection that supports
// VTODO, e.g. a Nextcloud Tasks list).
type Settings struct {
	ListURL  string
	Username string
	Password string
}

// NewSettings validates the settings.
func NewSettings(listURL, username, password string) (Settings, error) {
	listURL = strings.TrimSpace(listURL)
	u, err := url.Parse(listURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("listUrl", "The list URL must start with http:// or https://.")
	}
	if !strings.HasSuffix(listURL, "/") {
		listURL += "/"
	}
	if strings.TrimSpace(username) == "" || password == "" {
		return Settings{}, kernel.NewValidationError("username", "Username and app password are required.")
	}
	return Settings{ListURL: listURL, Username: strings.TrimSpace(username), Password: password}, nil
}

// ValidItemID reports whether id is a CalDAV object name inside the list.
func ValidItemID(id string) bool { return itemIDPattern.MatchString(id) && !strings.Contains(id, "..") }

// NormaliseSummary validates the text of a new item.
func NormaliseSummary(s string) (string, error) {
	s = strings.Join(strings.Fields(s), " ")
	if s == "" || utf8.RuneCountInString(s) > maxSummaryLength {
		return "", kernel.NewValidationError("summary", "Items must be between 1 and 200 characters long.")
	}
	return s, nil
}

// Item is one entry of the list.
type Item struct {
	ID        string
	Summary   string
	Done      bool
	Created   time.Time
	Completed time.Time
}

// ForDisplay returns open items (oldest first) followed by the most recently
// completed ones.
func ForDisplay(items []Item) []Item {
	var open, done []Item
	for _, it := range items {
		if it.Done {
			done = append(done, it)
		} else {
			open = append(open, it)
		}
	}
	slices.SortStableFunc(open, func(a, b Item) int { return cmp.Or(a.Created.Compare(b.Created), cmp.Compare(a.Summary, b.Summary)) })
	slices.SortStableFunc(done, func(a, b Item) int { return b.Completed.Compare(a.Completed) })
	if len(done) > MaxCompletedShown {
		done = done[:MaxCompletedShown]
	}
	return append(open, done...)
}
