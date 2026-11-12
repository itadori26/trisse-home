// Package domain holds speed measurements and threshold rules.
//
// Ring: Entities.
package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// ErrNotConfigured is returned while no Speedtest Tracker is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Speedtest Tracker is not configured yet.")

const maxThresholdMbps = 100000

// Settings are the connection details and the optional thresholds
// (0 disables a threshold).
type Settings struct {
	URL         string
	Token       string
	MinDownMbps float64
	MinUpMbps   float64
}

// NewSettings validates and normalises the settings.
func NewSettings(rawURL, token string, minDown, minUp float64) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like http://192.168.1.2:8765.")
	}
	if strings.TrimSpace(token) == "" {
		return Settings{}, kernel.NewValidationError("token", "An API token with the results:read scope is required.")
	}
	if minDown < 0 || minUp < 0 || minDown > maxThresholdMbps || minUp > maxThresholdMbps {
		return Settings{}, kernel.NewValidationError("minDownMbps", "Thresholds must be between 0 and 100000 Mbit/s.")
	}
	return Settings{URL: rawURL, Token: strings.TrimSpace(token), MinDownMbps: minDown, MinUpMbps: minUp}, nil
}

// HasThresholds reports whether notifications are wanted.
func (s Settings) HasThresholds() bool { return s.MinDownMbps > 0 || s.MinUpMbps > 0 }

// Measurement is one completed speed test.
type Measurement struct {
	At          time.Time
	DownloadBps float64 // bits per second
	UploadBps   float64
	PingMs      float64
}

// Mbps converts bits per second to Mbit/s.
func Mbps(bps float64) float64 { return bps / 1e6 }

// Judge checks a measurement against the thresholds; detail explains a
// violation.
func (s Settings) Judge(m Measurement) (healthy bool, detail string) {
	var problems []string
	if s.MinDownMbps > 0 && Mbps(m.DownloadBps) < s.MinDownMbps {
		problems = append(problems, fmt.Sprintf("download %.0f Mbit/s (minimum %.0f)", Mbps(m.DownloadBps), s.MinDownMbps))
	}
	if s.MinUpMbps > 0 && Mbps(m.UploadBps) < s.MinUpMbps {
		problems = append(problems, fmt.Sprintf("upload %.0f Mbit/s (minimum %.0f)", Mbps(m.UploadBps), s.MinUpMbps))
	}
	if len(problems) == 0 {
		return true, ""
	}
	return false, "Internet is slow: " + strings.Join(problems, ", ")
}
