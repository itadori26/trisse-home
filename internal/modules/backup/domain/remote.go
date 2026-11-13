package domain

import (
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
	u, err := kernel.ParseTarget("url", rawURL, kernel.WithTargetPath())
	if err != nil {
		return RemoteSettings{}, err
	}
	rawURL = u.String() + "/" // a WebDAV folder
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
