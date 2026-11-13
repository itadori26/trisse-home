package domain

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Provider is a media server the module can read from (#61).
type Provider string

// Supported providers.
const (
	ProviderJellyfin Provider = "jellyfin"
	ProviderPlex     Provider = "plex"
)

// ErrNotConfigured is returned while no media server is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "The media server is not configured yet.")

// ErrPosterNotFound is returned for an invalid or unknown item ID.
var ErrPosterNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The poster does not exist.")

// Settings are the connection details of the media server.
type Settings struct {
	Provider Provider
	URL      string
	// Token is the Jellyfin API key or the Plex token.
	Token string
}

// NewSettings validates and normalises the settings.
func NewSettings(provider, rawURL, token string) (Settings, error) {
	p := Provider(strings.TrimSpace(provider))
	if p != ProviderJellyfin && p != ProviderPlex {
		return Settings{}, kernel.NewValidationError("provider", "Choose Jellyfin or Plex.")
	}
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like http://192.168.1.2:8096.")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return Settings{}, kernel.NewValidationError("token", "The API key or token is required.")
	}
	return Settings{Provider: p, URL: rawURL, Token: token}, nil
}

// itemIDPattern matches Jellyfin GUIDs and Plex rating keys; anything else
// is never forwarded to the server.
var itemIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

// ValidItemID reports whether id may be used to request a poster.
func ValidItemID(id string) bool { return itemIDPattern.MatchString(id) }

// Item is a recently added movie, episode or album.
type Item struct {
	ID    string
	Title string
	// Series is the show of an episode or the artist of an album.
	Series  string
	Kind    string // movie, episode, series, album, …
	Year    int
	AddedAt time.Time
}

// Session is something that is being played right now.
type Session struct {
	User, Device string
	Title        string
	Series       string
	Kind         string
	Paused       bool
	Position     time.Duration
	Duration     time.Duration
}

// Progress returns the played share between 0 and 100.
func (s Session) Progress() float64 {
	if s.Duration <= 0 {
		return 0
	}
	return min(100, max(0, float64(s.Position)/float64(s.Duration)*100))
}

// Overview is what the widget shows.
type Overview struct {
	ServerName string
	Recent     []Item
	Sessions   []Session
}

// MaxRecent is how many recently added items are shown.
const MaxRecent = 12

// Poster is a cover image.
type Poster struct {
	ContentType string
	Data        []byte
}
