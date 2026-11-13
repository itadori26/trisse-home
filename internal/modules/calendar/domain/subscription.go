package domain

import (
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits of ICS subscriptions (#98).
const (
	MaxSubscriptions       = 20
	maxSubscriptionName    = 60
	DefaultRefreshInterval = time.Hour
	MinRefreshInterval     = 15 * time.Minute
	MaxRefreshInterval     = 24 * time.Hour
	// DefaultColor is used when no colour is chosen.
	DefaultColor = "#3b6ef5"
)

// ErrSubscriptionNotFound is returned for an unknown subscription ID.
var ErrSubscriptionNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The calendar subscription does not exist.")

// ErrTooManySubscriptions is returned when the limit is reached.
var ErrTooManySubscriptions = kernel.NewPublicErrorf(kernel.ErrInvalid, "At most {max} calendar subscriptions are supported.", "max", MaxSubscriptions)

var colorPattern = regexp.MustCompile(`^#[0-9a-f]{6}$`)

// Subscription is a read-only calendar loaded from an ICS URL, e.g. a
// university timetable or public holidays. The URL often contains a
// private token, so it is treated like a password.
type Subscription struct {
	ID              string
	Name            string
	URL             string
	Color           string
	RefreshInterval time.Duration
}

// NewSubscription validates a subscription. webcal:// links (as offered by
// many sites) are fetched over https://.
func NewSubscription(id, name, rawURL, color string, refresh time.Duration) (Subscription, error) {
	s := Subscription{ID: id, Name: strings.TrimSpace(name), Color: strings.ToLower(strings.TrimSpace(color)), RefreshInterval: refresh}
	if s.Name == "" || utf8.RuneCountInString(s.Name) > maxSubscriptionName {
		return Subscription{}, kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	rawURL = strings.TrimSpace(rawURL)
	if rest, ok := strings.CutPrefix(rawURL, "webcal://"); ok {
		rawURL = "https://" + rest
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Subscription{}, kernel.NewValidationError("url", "The ICS URL must start with https://, http:// or webcal://.")
	}
	s.URL = rawURL
	if s.Color == "" {
		s.Color = DefaultColor
	}
	if !colorPattern.MatchString(s.Color) {
		return Subscription{}, kernel.NewValidationError("color", "The colour must look like #3b6ef5.")
	}
	if s.RefreshInterval == 0 {
		s.RefreshInterval = DefaultRefreshInterval
	}
	if s.RefreshInterval < MinRefreshInterval || s.RefreshInterval > MaxRefreshInterval {
		return Subscription{}, kernel.NewValidationError("refreshMinutes", "The refresh interval must be between 15 minutes and 24 hours.")
	}
	return s, nil
}

// URLHint is a display form of the URL without path and query, which may
// carry a private token: "https://calendar.example.com/…".
func (s Subscription) URLHint() string {
	u, err := url.Parse(s.URL)
	if err != nil {
		return ""
	}
	hint := u.Scheme + "://" + u.Host
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" {
		hint += "/…"
	}
	return hint
}
