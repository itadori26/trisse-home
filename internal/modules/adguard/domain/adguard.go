package domain

import (
	"net/url"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// ErrNotConfigured is returned while no AdGuard instance is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "AdGuard Home is not configured yet.")

// Settings are the connection details of AdGuard Home.
type Settings struct {
	URL      string
	Username string
	Password string
}

// NewSettings validates and normalises the settings.
func NewSettings(rawURL, username, password string) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like http://192.168.1.2:3000.")
	}
	if strings.TrimSpace(username) == "" || password == "" {
		return Settings{}, kernel.NewValidationError("username", "Username and password are required.")
	}
	return Settings{URL: rawURL, Username: strings.TrimSpace(username), Password: password}, nil
}

// BlockedDomain is a domain with its number of queries.
type BlockedDomain struct {
	Domain string
	Count  int
}

// Stats is a snapshot of AdGuard Home statistics.
type Stats struct {
	Queries           int
	Blocked           int
	AvgProcessingTime time.Duration
	ProtectionEnabled bool
	TopBlocked        []BlockedDomain
}

// BlockedPercent returns the share of blocked queries.
func (s Stats) BlockedPercent() float64 {
	if s.Queries == 0 {
		return 0
	}
	return float64(s.Blocked) / float64(s.Queries) * 100
}
