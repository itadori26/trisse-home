package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// ErrDisabled is returned while a source is switched off in the settings.
var ErrDisabled = kernel.NewPublicError(kernel.ErrNotFound, "This space source is switched off in the settings.")

// LaunchStatus is the normalised status of a launch.
type LaunchStatus string

// Launch statuses (Launch Library 2 abbreviations, normalised).
const (
	StatusGo       LaunchStatus = "go"
	StatusTBD      LaunchStatus = "tbd"
	StatusTBC      LaunchStatus = "tbc"
	StatusHold     LaunchStatus = "hold"
	StatusInFlight LaunchStatus = "in_flight"
	StatusSuccess  LaunchStatus = "success"
	StatusFailure  LaunchStatus = "failure"
	StatusPartial  LaunchStatus = "partial_failure"
	StatusUnknown  LaunchStatus = "unknown"
)

// ParseLaunchStatus maps an LL2 abbreviation ("Go", "TBD", "In Flight", …).
func ParseLaunchStatus(abbrev string) LaunchStatus {
	switch strings.ToLower(strings.TrimSpace(abbrev)) {
	case "go":
		return StatusGo
	case "tbd":
		return StatusTBD
	case "tbc":
		return StatusTBC
	case "hold":
		return StatusHold
	case "in flight":
		return StatusInFlight
	case "success":
		return StatusSuccess
	case "failure":
		return StatusFailure
	case "partial failure":
		return StatusPartial
	default:
		return StatusUnknown
	}
}

// Launch is an upcoming launch.
type Launch struct {
	ID          string
	Mission     string
	MissionType string
	Description string
	Rocket      string
	Provider    string
	Pad         string
	Location    string
	// Net is the "no earlier than" time.
	Net         time.Time
	Status      LaunchStatus
	WebcastLive bool
}

// MaxDescription bounds the mission description kept per launch.
const MaxDescription = 600

// Truncate shortens s to n runes with an ellipsis.
func Truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:n-1])) + "…"
}

// MaxLaunches bounds the list.
const MaxLaunches = 10

// Picture is the Astronomy Picture of the Day.
type Picture struct {
	Date        string
	Title       string
	Explanation string
	Copyright   string
	// MediaType is "image", "video" or something else.
	MediaType string
	// URL is the image (or video page); HDURL the large image.
	URL, HDURL string
}

// IsImage reports whether the picture is an image the widget can show.
func (p Picture) IsImage() bool { return p.MediaType == "image" && p.URL != "" }

// Settings switch the sources on; both are off until enabled. The NASA API
// key is a secret; without one the public DEMO_KEY (low rate limit) is used.
type Settings struct {
	Launches bool
	Picture  bool
	NASAKey  string
}

// HasKey reports whether an own NASA API key is stored.
func (s Settings) HasKey() bool { return s.NASAKey != "" }

// NewSettings validates the settings; an empty key with keepKey keeps stored.
func NewSettings(launches, picture bool, key string, keepKey bool, stored Settings) (Settings, error) {
	key = strings.TrimSpace(key)
	if key == "" && keepKey {
		key = stored.NASAKey
	}
	if len(key) > 100 || strings.ContainsAny(key, " \r\n\t&?#/") {
		return Settings{}, kernel.NewValidationError("nasaKey", "The NASA API key looks wrong.")
	}
	return Settings{Launches: launches, Picture: picture, NASAKey: key}, nil
}
