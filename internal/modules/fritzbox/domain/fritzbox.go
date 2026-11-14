package domain

import (
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// DefaultURL is the TR-064 endpoint of a FRITZ!Box in its default network.
const DefaultURL = "http://192.168.178.1:49000"

// MaxCalls is the most calls a calls widget shows (#168); the box is asked
// for this many and the widgets filter them.
const MaxCalls = 50

// DefaultCalls is the number of calls shown unless the widget says otherwise.
const DefaultCalls = 10

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

// Status is a snapshot of the internet connection. The call list is not part
// of it (#168): it is only fetched when a calls widget asks for it.
type Status struct {
	Model string
	WAN   WAN
	// Hint explains missing data (UPnP status disabled …).
	Hint string
}

// CallKind selects the calls a calls widget shows.
type CallKind string

// Call kinds of a CallFilter.
const (
	CallsAll      CallKind = "all"
	CallsMissed   CallKind = "missed"
	CallsIncoming CallKind = "incoming"
	CallsOutgoing CallKind = "outgoing"
)

// CallFilter is what a calls widget shows: a kind of call and how many.
type CallFilter struct {
	kind  CallKind
	limit int
}

// NewCallFilter validates the filter; an empty kind means all calls and a
// limit of 0 the default.
func NewCallFilter(kind string, limit int) (CallFilter, error) {
	k := CallKind(strings.TrimSpace(kind))
	if k == "" {
		k = CallsAll
	}
	switch k {
	case CallsAll, CallsMissed, CallsIncoming, CallsOutgoing:
	default:
		return CallFilter{}, kernel.NewValidationError("filter", "The filter must be all, missed, incoming or outgoing.")
	}
	if limit == 0 {
		limit = DefaultCalls
	}
	if limit < 5 || limit > MaxCalls {
		return CallFilter{}, kernel.NewValidationErrorf("limit", "Show between {min} and {max} calls.", "min", 5, "max", MaxCalls)
	}
	return CallFilter{kind: k, limit: limit}, nil
}

// Kind returns the selected kind of call.
func (f CallFilter) Kind() CallKind { return f.kind }

// Limit returns the number of calls shown.
func (f CallFilter) Limit() int { return f.limit }

// Apply keeps the calls of the selected kind, in the given order, at most Limit.
func (f CallFilter) Apply(calls []Call) []Call {
	out := make([]Call, 0, min(len(calls), f.limit))
	for _, c := range calls {
		if len(out) == f.limit {
			break
		}
		if f.kind == CallsAll || string(c.Type) == string(f.kind) {
			out = append(out, c)
		}
	}
	return out
}
