package domain

import (
	"net/url"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// DefaultRemoteRetention is the number of backups kept off-site (#46).
const DefaultRemoteRetention = 30

// ErrRemoteNotConfigured is returned while no off-site target is set.
var ErrRemoteNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "No off-site backup target is configured.")

// RemoteSettings describe a WebDAV target such as a Nextcloud folder.
type RemoteSettings struct {
	URL       string // collection URL, always ending with "/"
	Username  string
	Password  string
	Retention int // backups kept remotely, independent of local retention
}

// NewRemoteSettings validates and normalises the settings.
func NewRemoteSettings(rawURL, username, password string, retention int) (RemoteSettings, error) {
	rawURL = strings.TrimSpace(rawURL)
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
		return RemoteSettings{}, kernel.NewValidationError("url",
			"The WebDAV URL must look like https://cloud.example.org/remote.php/dav/files/USER/Backups/.")
	}
	if !strings.HasSuffix(rawURL, "/") {
		rawURL += "/"
	}
	if strings.TrimSpace(username) == "" || password == "" {
		return RemoteSettings{}, kernel.NewValidationError("username", "User name and password (or app password) are required.")
	}
	if retention < 1 || retention > MaxRetention {
		return RemoteSettings{}, kernel.NewValidationError("retention", "Between 1 and 365 backups can be kept.")
	}
	return RemoteSettings{URL: rawURL, Username: strings.TrimSpace(username), Password: password, Retention: retention}, nil
}

// RemotePolicy returns the retention policy of the off-site copies.
func (s RemoteSettings) RemotePolicy() Policy {
	return Policy{Interval: MinInterval, Retention: s.Retention}
}
