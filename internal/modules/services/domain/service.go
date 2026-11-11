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
	// MAC is the hardware address for Wake-on-LAN (optional).
	MAC kernel.MAC
}

// ErrNoMAC is returned when waking a service without MAC address.
var ErrNoMAC = kernel.NewPublicError(kernel.ErrInvalid, "No MAC address is stored for this service.")

// SetMAC validates and stores the Wake-on-LAN address; empty removes it.
func (s *Service) SetMAC(raw string) error {
	m, err := kernel.ParseMAC(raw)
	if err != nil {
		return err
	}
	s.MAC = m
	return nil
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
	// CertExpiresAt is the expiry of the TLS certificate (zero for HTTP).
	CertExpiresAt time.Time
}

// CheckResult is the raw outcome of one HTTP health check.
type CheckResult struct {
	HTTPStatus    int
	Latency       time.Duration
	CertExpiresAt time.Time
	Err           error
}

// DefaultCertWarning is how long before expiry a certificate is flagged.
const DefaultCertWarning = 14 * 24 * time.Hour

// CertificateExpiresWithin reports whether the TLS certificate expires within d of now.
func (s Status) CertificateExpiresWithin(d time.Duration, now time.Time) bool {
	return !s.CertExpiresAt.IsZero() && s.CertExpiresAt.Sub(now) < d
}

// StatusFromCheck classifies a check: any HTTP answer below 500 means the
// service is up (401/403 are typical for protected services).
func StatusFromCheck(serviceID string, r CheckResult, at time.Time) Status {
	s := StatusFromResponse(serviceID, r.HTTPStatus, r.Latency, r.Err, at)
	s.CertExpiresAt = r.CertExpiresAt
	return s
}

// StatusFromResponse classifies a check without TLS information.
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
