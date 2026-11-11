package domain

import (
	"net/url"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// ErrNotConfigured is returned while no Nextcloud instance is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Nextcloud is not configured yet.")

// Settings are the connection details: base URL, user and app password.
type Settings struct {
	URL         string
	Username    string
	AppPassword string
}

// NewSettings validates and normalises the settings.
func NewSettings(rawURL, username, appPassword string) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like https://cloud.example.com.")
	}
	if strings.TrimSpace(username) == "" || appPassword == "" {
		return Settings{}, kernel.NewValidationError("username", "Username and app password are required.")
	}
	return Settings{URL: rawURL, Username: strings.TrimSpace(username), AppPassword: appPassword}, nil
}

// Quota is the storage of the configured account in bytes. Total is 0 when
// the quota is unlimited.
type Quota struct {
	Used  int64
	Free  int64
	Total int64
}

// Percent returns the used share, or 0 for unlimited quotas.
func (q Quota) Percent() float64 {
	if q.Total <= 0 {
		return 0
	}
	return float64(q.Used) / float64(q.Total) * 100
}

// ServerInfo is provided by the serverinfo app (admins only).
type ServerInfo struct {
	FreeSpace     int64
	Users         int
	ActiveUsers5m int
	ActiveUsers1d int
	Files         int
	AppUpdates    int
	DatabaseSize  int64
}

// Status is a snapshot of the instance.
type Status struct {
	Version     string
	Maintenance bool
	Quota       Quota
	// Server is nil when serverinfo is unavailable; ServerHint explains why.
	Server     *ServerInfo
	ServerHint string
}

// ServerInfoHint explains a failed serverinfo request by HTTP status.
func ServerInfoHint(httpStatus int) string {
	switch httpStatus {
	case 401, 403, 997:
		return "Server details need an admin account (or the serverinfo app is disabled)."
	case 404, 998:
		return "Server details need the serverinfo app (Apps → Monitoring)."
	default:
		return "Server details are not available."
	}
}
