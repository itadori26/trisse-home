package domain

import (
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Provider is a DNS filter the module can read statistics from (#83).
type Provider string

// Supported providers.
const (
	ProviderAdGuard Provider = "adguard"
	ProviderPiHole  Provider = "pihole"
)

// Providers lists every supported provider.
var Providers = []Provider{ProviderAdGuard, ProviderPiHole}

// ParseProvider maps a request value to a provider; empty means AdGuard
// Home, which keeps widgets and API clients from before #83 working.
func ParseProvider(raw string) (Provider, error) {
	switch p := Provider(strings.TrimSpace(raw)); p {
	case "":
		return ProviderAdGuard, nil
	case ProviderAdGuard, ProviderPiHole:
		return p, nil
	}
	return "", kernel.NewValidationError("provider", "Unknown DNS filter, use adguard or pihole.")
}

// ErrNotConfigured is returned while the provider is not configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "This DNS filter is not configured yet.")

// Settings are the connection details of a DNS filter.
type Settings struct {
	Provider Provider
	URL      string
	// Username is only used by AdGuard Home.
	Username string
	// Password is the AdGuard Home password or the Pi-hole (app) password;
	// Pi-hole without a password needs none.
	Password string
}

// NewSettings validates and normalises the settings of a provider.
func NewSettings(p Provider, rawURL, username, password string) (Settings, error) {
	// A path is allowed: Pi-hole and reverse proxies often serve a sub path.
	u, err := kernel.ParseTarget("url", rawURL, kernel.WithTargetPath())
	if err != nil {
		return Settings{}, err
	}
	s := Settings{Provider: p, URL: u.String(), Password: password}
	switch p {
	case ProviderAdGuard:
		if strings.TrimSpace(username) == "" || password == "" {
			return Settings{}, kernel.NewValidationError("username", "Username and password are required.")
		}
		s.Username = strings.TrimSpace(username)
	case ProviderPiHole:
		// Pi-hole v6 authenticates with a password only.
	default:
		return Settings{}, kernel.NewValidationError("provider", "Unknown DNS filter, use adguard or pihole.")
	}
	return s, nil
}

// BlockedDomain is a domain with its number of queries.
type BlockedDomain struct {
	Domain string
	Count  int
}

// Stats is a provider-neutral snapshot of DNS filter statistics.
type Stats struct {
	Queries int
	Blocked int
	// AvgProcessingTime is zero when the provider does not report it
	// (Pi-hole); HasAvgProcessingTime tells both cases apart.
	AvgProcessingTime    time.Duration
	HasAvgProcessingTime bool
	ProtectionEnabled    bool
	// PauseRemaining is the time left of a timed pause; zero when the
	// filter is on, or off without a timer.
	PauseRemaining time.Duration
	TopBlocked     []BlockedDomain
}

// BlockedPercent returns the share of blocked queries.
func (s Stats) BlockedPercent() float64 {
	if s.Queries == 0 {
		return 0
	}
	return float64(s.Blocked) / float64(s.Queries) * 100
}

// TopBlockedLimit is the number of top blocked domains shown.
const TopBlockedLimit = 5

// Pausing the filter (#177) is an opt-in write action: the providers that
// may be paused come from the server environment, never from a session.

// PauseMinutes are the offered pause durations.
var PauseMinutes = []int{5, 30, 60}

// ParsePause validates a pause duration in minutes.
func ParsePause(minutes int) (time.Duration, error) {
	for _, m := range PauseMinutes {
		if m == minutes {
			return time.Duration(m) * time.Minute, nil
		}
	}
	return 0, kernel.NewValidationError("minutes", "Pause for 5, 30 or 60 minutes.")
}

// ErrPauseNotAllowed is returned when pausing is not enabled for a provider.
var ErrPauseNotAllowed = kernel.NewPublicError(kernel.ErrForbidden,
	"Pausing is switched off for this DNS filter. Allow it with DNS_PAUSE_PROVIDERS on the server.")
