// Package domain holds the Service entity and its health Status.
//
// Ring: Entities.
package domain

import (
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Field limits.
const (
	MaxNameLength        = 60
	MaxDescriptionLength = 200
	MaxIconLength        = 2048
)

// ErrServiceNotFound is returned for unknown service IDs.
var ErrServiceNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The service does not exist.")

// Service is a tile linking to a self-hosted service.
type Service struct {
	ID            string
	Name          string
	URL           string
	CheckURL      string
	Icon          string
	Description   string
	SkipTLSVerify bool
	Position      int
}

// NewService validates all fields. CheckURL defaults to URL; Icon may be an
// emoji/short text or an http(s) image URL.
func NewService(id, name, rawURL, checkURL, icon, description string, skipTLS bool, position int) (Service, error) {
	s := Service{
		ID: id, Name: strings.TrimSpace(name), URL: strings.TrimSpace(rawURL), CheckURL: strings.TrimSpace(checkURL),
		Icon: strings.TrimSpace(icon), Description: strings.TrimSpace(description), SkipTLSVerify: skipTLS, Position: position,
	}
	if s.Name == "" || utf8.RuneCountInString(s.Name) > MaxNameLength {
		return Service{}, kernel.NewValidationError("name", "The name must be between 1 and 60 characters long.")
	}
	if !isHTTPURL(s.URL) {
		return Service{}, kernel.NewValidationError("url", "The URL must start with http:// or https://.")
	}
	if s.CheckURL == "" {
		s.CheckURL = s.URL
	}
	if !isHTTPURL(s.CheckURL) {
		return Service{}, kernel.NewValidationError("checkUrl", "The check URL must start with http:// or https://.")
	}
	if len(s.Icon) > MaxIconLength {
		return Service{}, kernel.NewValidationError("icon", "The icon is too long.")
	}
	if utf8.RuneCountInString(s.Description) > MaxDescriptionLength {
		return Service{}, kernel.NewValidationError("description", "The description must be at most 200 characters long.")
	}
	return s, nil
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// State is the result of a health check.
type State string

// Health states.
const (
	StateUnknown State = "unknown"
	StateOnline  State = "online"
	StateOffline State = "offline"
)

// Status is the latest health check result of a service.
type Status struct {
	ServiceID  string
	State      State
	HTTPStatus int
	Latency    time.Duration
	CheckedAt  time.Time
	Error      string
}

// StatusFromResponse classifies a check: any HTTP answer below 500 means the
// service is up (401/403 are typical for protected services).
func StatusFromResponse(serviceID string, httpStatus int, latency time.Duration, checkErr error, at time.Time) Status {
	s := Status{ServiceID: serviceID, HTTPStatus: httpStatus, Latency: latency, CheckedAt: at, State: StateOnline}
	switch {
	case checkErr != nil:
		s.State, s.Error = StateOffline, checkErr.Error()
	case httpStatus >= 500:
		s.State = StateOffline
	}
	return s
}
