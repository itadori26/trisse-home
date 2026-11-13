package domain

import (
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// DefaultURL is the TR-064 endpoint of a FRITZ!Box in its default network.
const DefaultURL = "http://192.168.178.1:49000"

// MaxCalls is the number of calls shown.
const MaxCalls = 10

// ErrNotConfigured is returned while no FRITZ!Box is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "The FRITZ!Box is not configured yet.")

// Settings are the connection details. Username and Password are optional;
// without them only the UPnP status is available.
type Settings struct {
	URL      string
	Username string
	Password string
}

// HasLogin reports whether TR-064 credentials are configured.
func (s Settings) HasLogin() bool { return s.Password != "" }

// NewSettings validates the settings. The URL must point to the TR-064 port
// (usually 49000); a missing port is added.
func NewSettings(rawURL, username, password string) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	if rawURL == "" {
		rawURL = DefaultURL
	}
	u, err := kernel.ParseTarget("url", rawURL)
	if err != nil {
		return Settings{}, err
	}
	// TR-064 on port 49000 only speaks http; https uses a different port.
	if u.Scheme != "http" {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like http://192.168.178.1:49000.")
	}
	if u.Port() == "" {
		u.Host += ":49000"
	}
	username = strings.TrimSpace(username)
	if (username == "") != (password == "") && password == "" {
		return Settings{}, kernel.NewValidationError("password", "Please enter the password of the FRITZ!Box user.")
	}
	return Settings{URL: u.Scheme + "://" + u.Host, Username: username, Password: password}, nil
}

// WAN is the state of the internet connection.
type WAN struct {
	Connected  bool
	ExternalIP string
	Uptime     time.Duration
	// Current throughput in bytes per second; Known is false when the box
	// does not report it (UPnP status disabled).
	Known        bool
	DownstreamBs int64
	UpstreamBs   int64
	// Link speed in bits per second (0 = unknown, needs a login).
	DownstreamMaxBits int64
	UpstreamMaxBits   int64
}

// DownstreamPercent is the current download utilisation (0 if unknown).
func (w WAN) DownstreamPercent() float64 { return percent(w.DownstreamBs, w.DownstreamMaxBits) }

// UpstreamPercent is the current upload utilisation (0 if unknown).
func (w WAN) UpstreamPercent() float64 { return percent(w.UpstreamBs, w.UpstreamMaxBits) }

func percent(bytesPerSecond, maxBits int64) float64 {
	if maxBits <= 0 {
		return 0
	}
	return min(100, float64(bytesPerSecond*8)/float64(maxBits)*100)
}

// CallType classifies a call.
type CallType string

// Call types.
const (
	CallIncoming CallType = "incoming"
	CallMissed   CallType = "missed"
	CallOutgoing CallType = "outgoing"
	CallRejected CallType = "rejected"
	CallActive   CallType = "active"
)

// CallTypeFromAVM maps the numeric type of the AVM call list.
func CallTypeFromAVM(t int) CallType {
	switch t {
	case 1:
		return CallIncoming
	case 2:
		return CallMissed
	case 3:
		return CallOutgoing
	case 10:
		return CallRejected
	default:
		return CallActive
	}
}

// Call is one entry of the call list.
type Call struct {
	Type     CallType
	Number   string
	Name     string
	At       time.Time
	Duration time.Duration
}

// Status is a snapshot of the box.
type Status struct {
	Model string
	WAN   WAN
	Calls []Call
	// Hint explains missing data (no login, UPnP status disabled …).
	Hint string
}
