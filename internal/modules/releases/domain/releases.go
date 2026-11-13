package domain

import (
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits of the module (#63).
const (
	MaxFollowed  = 50
	DefaultWeeks = 4
	MaxWeeks     = 12
	// MinQueryLength avoids searching for single letters.
	MinQueryLength = 2
)

// ErrTooManyArtists is returned when the follow list is full.
var ErrTooManyArtists = kernel.NewPublicErrorf(kernel.ErrInvalid, "At most {max} artists can be followed.", "max", MaxFollowed)

// ErrCoverNotFound is returned for a release that is not in the list.
var ErrCoverNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The cover does not exist.")

// mbidPattern matches MusicBrainz identifiers (UUIDs).
var mbidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidMBID reports whether id is a MusicBrainz identifier.
func ValidMBID(id string) bool { return mbidPattern.MatchString(id) }

// Artist is a MusicBrainz artist.
type Artist struct {
	ID   string
	Name string
	// Disambiguation tells artists with the same name apart, e.g. "UK band".
	Disambiguation string
}

// NewArtist validates an artist to follow.
func NewArtist(id, name string) (Artist, error) {
	name = strings.TrimSpace(name)
	if !ValidMBID(id) {
		return Artist{}, kernel.NewValidationError("id", "The artist ID is not a MusicBrainz ID.")
	}
	if name == "" || utf8.RuneCountInString(name) > 200 {
		return Artist{}, kernel.NewValidationError("name", "The artist name must be between 1 and 200 characters long.")
	}
	return Artist{ID: id, Name: name}, nil
}

// Release is a release group (album, single, EP) of a followed artist.
type Release struct {
	ID     string
	Title  string
	Artist string
	Type   string // Album, Single, EP, …
	Date   time.Time
}

// ClampWeeks bounds the period of the release list.
func ClampWeeks(weeks int) int {
	if weeks <= 0 {
		return DefaultWeeks
	}
	return min(weeks, MaxWeeks)
}

// Recent returns the releases since now minus weeks, newest first.
func Recent(list []Release, weeks int, now time.Time) []Release {
	since := now.AddDate(0, 0, -7*ClampWeeks(weeks))
	out := make([]Release, 0, len(list))
	for _, r := range list {
		if !r.Date.Before(since) && !r.Date.After(now) {
			out = append(out, r)
		}
	}
	slices.SortStableFunc(out, func(a, b Release) int {
		if c := b.Date.Compare(a.Date); c != 0 {
			return c
		}
		return strings.Compare(a.Title, b.Title)
	})
	return out
}

// ParseDate reads MusicBrainz dates, which may lack month or day.
func ParseDate(s string) (time.Time, bool) {
	for _, layout := range []string{time.DateOnly, "2006-01", "2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Cover is a cover image.
type Cover struct {
	ContentType string
	Data        []byte
}
