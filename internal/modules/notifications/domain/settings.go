package domain

import (
	"net/url"
	"regexp"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// Rule limits and defaults.
const (
	DefaultServerURL     = "https://ntfy.sh"
	DefaultDiskThreshold = 90
	DefaultTLSDays       = 14
	minDiskThreshold     = 50
	maxDiskThreshold     = 99
	maxTLSDays           = 90
)

var topicPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// Rules select what is notified.
type Rules struct {
	Services      bool
	Containers    bool
	Disk          bool
	DiskThreshold float64
	TLS           bool
	TLSDays       int
	Backups       bool
	// Security delivers security alerts (new sign-ins, lockouts).
	Security bool
}

// DefaultRules enable everything with sensible thresholds.
func DefaultRules() Rules {
	return Rules{Services: true, Containers: true, Disk: true, DiskThreshold: DefaultDiskThreshold, TLS: true, TLSDays: DefaultTLSDays, Backups: true, Security: true}
}

// Settings configure delivery via ntfy.
type Settings struct {
	Enabled   bool
	ServerURL string
	Topic     string
	Token     string
	Rules     Rules
	// Policy holds severities and quiet hours (#69).
	Policy Policy
}

// DefaultSettings are used until the user configures notifications.
func DefaultSettings() Settings {
	return Settings{ServerURL: DefaultServerURL, Rules: DefaultRules()}
}

// NewSettings validates settings. A topic is only required when enabled.
func NewSettings(enabled bool, serverURL, topic, token string, rules Rules) (Settings, error) {
	serverURL = strings.TrimRight(strings.TrimSpace(serverURL), "/")
	if serverURL == "" {
		serverURL = DefaultServerURL
	}
	u, err := url.Parse(serverURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("serverUrl", "The ntfy server URL must start with http:// or https://.")
	}
	topic = strings.TrimSpace(topic)
	if (enabled || topic != "") && !topicPattern.MatchString(topic) {
		return Settings{}, kernel.NewValidationError("topic", "The topic may contain letters, digits, - and _ (max. 64 characters).")
	}
	if rules.DiskThreshold < minDiskThreshold || rules.DiskThreshold > maxDiskThreshold {
		return Settings{}, kernel.NewValidationError("diskThreshold", "The disk threshold must be between 50 and 99 %.")
	}
	if rules.TLSDays < 1 || rules.TLSDays > maxTLSDays {
		return Settings{}, kernel.NewValidationError("tlsDays", "The certificate warning must be between 1 and 90 days.")
	}
	return Settings{Enabled: enabled, ServerURL: serverURL, Topic: topic, Token: strings.TrimSpace(token), Rules: rules}, nil
}

// Judge decides whether an observation is covered by the rules and whether
// it counts as healthy. Threshold kinds are judged by their value.
func (r Rules) Judge(o kernel.Observation) (relevant, healthy bool) {
	switch o.Kind {
	case "service":
		return r.Services, o.Healthy
	case "container":
		return r.Containers, o.Healthy
	case "backup", "check":
		return r.Backups, o.Healthy
	case "disk":
		return r.Disk, o.Value < r.DiskThreshold
	case "diskhealth", "forecast":
		return r.Disk, o.Healthy
	case "speedtest":
		// Opt-in by setting thresholds in the speedtest module.
		return true, o.Healthy
	case "tls":
		return r.TLS, o.Value > float64(r.TLSDays)
	default:
		return false, true
	}
}
