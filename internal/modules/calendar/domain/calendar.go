package domain

import (
	"net/url"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Limits of the calendar module.
const (
	MaxCalendars = 10
	// Horizon is how far ahead events are loaded.
	Horizon = 31 * 24 * time.Hour
	// MaxRange is the longest period queried at once (a six-week month grid).
	MaxRange          = 42 * 24 * time.Hour
	maxSummaryLength  = 200
	maxLocationLength = 200
)

// ErrRange is returned for an invalid or too long query period.
var ErrRange = kernel.NewValidationError("to", "The period must be positive and at most 42 days long.")

// ErrUnknownCalendar is returned when an event should be written to a
// calendar that is not configured.
var ErrUnknownCalendar = kernel.NewValidationError("calendarUrl", "Please choose one of the configured calendars.")

// CheckRange validates a query period.
func CheckRange(from, to time.Time) error {
	if !to.After(from) || to.Sub(from) > MaxRange {
		return ErrRange
	}
	return nil
}

// Contains reports whether url is one of the configured calendars.
func (s Settings) Contains(calendarURL string) bool {
	if !strings.HasSuffix(calendarURL, "/") {
		calendarURL += "/"
	}
	return slices.Contains(s.CalendarURLs, calendarURL)
}

// NewEvent validates a new event. For all-day events start and end are
// dates (midnight UTC) and end is exclusive; timed events must end after
// they start.
func NewEvent(uid, summary, location string, start, end time.Time, allDay bool) (Event, error) {
	e := Event{UID: uid, Summary: strings.TrimSpace(summary), Location: strings.TrimSpace(location), AllDay: allDay}
	if e.Summary == "" || len([]rune(e.Summary)) > maxSummaryLength {
		return Event{}, kernel.NewValidationError("summary", "The title must be between 1 and 200 characters long.")
	}
	if len([]rune(e.Location)) > maxLocationLength {
		return Event{}, kernel.NewValidationError("location", "The location must be at most 200 characters long.")
	}
	if allDay {
		start, end = midnightUTC(start), midnightUTC(end)
	}
	if !end.After(start) {
		return Event{}, kernel.NewValidationError("end", "The event must end after it starts.")
	}
	if end.Sub(start) > 366*24*time.Hour {
		return Event{}, kernel.NewValidationError("end", "Events can be at most one year long.")
	}
	e.Start, e.End = start, end
	return e, nil
}

func midnightUTC(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ErrNotConfigured is returned while no calendar is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "No calendar is configured yet.")

// Settings are the CalDAV connection details. Each URL points to one
// calendar collection, e.g.
// https://cloud.example.com/remote.php/dav/calendars/USER/personal/.
type Settings struct {
	CalendarURLs []string
	Username     string
	Password     string
}

// NewSettings validates the settings.
func NewSettings(urls []string, username, password string) (Settings, error) {
	var clean []string
	for _, raw := range urls {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return Settings{}, kernel.NewValidationError("calendarUrls", "Every calendar URL must start with http:// or https://.")
		}
		if !strings.HasSuffix(raw, "/") {
			raw += "/"
		}
		if !slices.Contains(clean, raw) {
			clean = append(clean, raw)
		}
	}
	switch {
	case len(clean) == 0:
		return Settings{}, kernel.NewValidationError("calendarUrls", "Please enter at least one calendar URL.")
	case len(clean) > MaxCalendars:
		return Settings{}, kernel.NewValidationError("calendarUrls", "At most 10 calendars are supported.")
	case strings.TrimSpace(username) == "" || password == "":
		return Settings{}, kernel.NewValidationError("username", "Username and app password are required.")
	}
	return Settings{CalendarURLs: clean, Username: strings.TrimSpace(username), Password: password}, nil
}

// CalendarName derives a display name from a collection URL ("personal").
func CalendarName(collectionURL string) string {
	u, err := url.Parse(collectionURL)
	if err != nil {
		return collectionURL
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	name, _ := url.PathUnescape(parts[len(parts)-1])
	if name == "" {
		return u.Host
	}
	return name
}

// Event is one (possibly expanded recurring) calendar event. All-day events
// have AllDay set and Start/End at midnight UTC of their dates; End is
// exclusive.
type Event struct {
	UID      string
	Summary  string
	Location string
	Start    time.Time
	End      time.Time
	AllDay   bool
	Calendar string
}

// EndsAfter reports whether the event is still relevant at t.
func (e Event) EndsAfter(t time.Time) bool {
	if e.AllDay {
		y, m, d := t.Date()
		today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
		return e.End.After(today)
	}
	return e.End.After(t)
}

// SortEvents orders events by start; all-day events come first on a day.
func SortEvents(events []Event) {
	slices.SortStableFunc(events, func(a, b Event) int {
		ad, bd := a.Start.Format(time.DateOnly), b.Start.Format(time.DateOnly)
		if ad != bd {
			return strings.Compare(ad, bd)
		}
		if a.AllDay != b.AllDay {
			if a.AllDay {
				return -1
			}
			return 1
		}
		return a.Start.Compare(b.Start)
	})
}
