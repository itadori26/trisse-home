package domain

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits.
const (
	MaxEndpoints      = 50
	MaxExtractions    = 10
	MinInterval       = 30 * time.Second
	MaxInterval       = 24 * time.Hour
	DefaultInterval   = 5 * time.Minute
	maxNameRunes      = 60
	maxLabelRunes     = 40
	maxPathRunes      = 200
	maxUnitRunes      = 12
	maxHeaderValueLen = 2048
	maxTextRunes      = 80
)

// ErrNotFound is returned for an unknown endpoint.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The endpoint does not exist.")

var headerName = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+.^_|~-]{1,64}$`)

// forbiddenHeaders are set by the HTTP client itself.
var forbiddenHeaders = []string{"host", "content-length", "transfer-encoding", "connection", "te", "upgrade", "trailer"}

// Extraction picks one value out of the response.
type Extraction struct {
	Label string
	// Path is a JSONPath such as $.data.temperature or $.items[0]['name'].
	Path string
	Unit string
	// Warn and Crit are optional thresholds; Below turns "at or above" into
	// "at or below" (e.g. free disk space).
	Warn, Crit *float64
	Below      bool
}

// Endpoint is a configured JSON endpoint.
type Endpoint struct {
	ID   string
	Name string
	URL  string
	// HeaderName/HeaderValue is one optional request header, e.g. an API key.
	// The value is a secret: stored encrypted, never returned by the API.
	HeaderName  string
	HeaderValue string
	Interval    time.Duration
	// AllowPublic permits addresses outside loopback and private ranges.
	AllowPublic bool
	Extractions []Extraction
}

// HasHeader reports whether a header is configured.
func (e Endpoint) HasHeader() bool { return e.HeaderName != "" }

// EndpointInput is the editable form; an empty HeaderValue with KeepHeader
// keeps the stored secret.
type EndpointInput struct {
	Name        string
	URL         string
	HeaderName  string
	HeaderValue string
	KeepHeader  bool
	Interval    time.Duration
	AllowPublic bool
	Extractions []Extraction
}

// PathCheck validates JSONPath syntax (implemented outside the domain).
type PathCheck func(path string) error

// NewEndpoint validates input. stored is the existing endpoint when
// updating (its header secret is kept if asked to), or nil.
func NewEndpoint(id string, in EndpointInput, stored *Endpoint, checkPath PathCheck) (Endpoint, error) {
	e := Endpoint{ID: id, Name: strings.TrimSpace(in.Name), Interval: in.Interval, AllowPublic: in.AllowPublic}
	if e.Name == "" || utf8.RuneCountInString(e.Name) > maxNameRunes {
		return Endpoint{}, kernel.NewValidationError("name", "Enter a name of at most 60 characters.")
	}
	u, err := kernel.ParseTarget("url", in.URL, kernel.WithTargetQuery())
	if err != nil {
		return Endpoint{}, err
	}
	e.URL = u.String()
	if e.Interval == 0 {
		e.Interval = DefaultInterval
	}
	if e.Interval < MinInterval || e.Interval > MaxInterval {
		return Endpoint{}, kernel.NewValidationError("interval", "The interval must be between 30 seconds and 24 hours.")
	}
	if err := e.setHeader(in, stored); err != nil {
		return Endpoint{}, err
	}
	if len(in.Extractions) == 0 || len(in.Extractions) > MaxExtractions {
		return Endpoint{}, kernel.NewValidationError("extractions", "Add between 1 and 10 values.")
	}
	for _, x := range in.Extractions {
		x.Label, x.Path, x.Unit = strings.TrimSpace(x.Label), strings.TrimSpace(x.Path), strings.TrimSpace(x.Unit)
		switch {
		case x.Label == "" || utf8.RuneCountInString(x.Label) > maxLabelRunes:
			return Endpoint{}, kernel.NewValidationError("extractions", "Every value needs a label of at most 40 characters.")
		case x.Path == "" || utf8.RuneCountInString(x.Path) > maxPathRunes:
			return Endpoint{}, kernel.NewValidationError("extractions", "Every value needs a JSONPath of at most 200 characters.")
		case utf8.RuneCountInString(x.Unit) > maxUnitRunes:
			return Endpoint{}, kernel.NewValidationError("extractions", "Units must be at most 12 characters long.")
		case !finite(x.Warn) || !finite(x.Crit):
			return Endpoint{}, kernel.NewValidationError("extractions", "Thresholds must be numbers.")
		}
		if err := checkPath(x.Path); err != nil {
			return Endpoint{}, kernel.NewValidationErrorf("extractions", "Invalid JSONPath {path}: {detail}", "path", x.Path, "detail", err.Error())
		}
		e.Extractions = append(e.Extractions, x)
	}
	return e, nil
}

func (e *Endpoint) setHeader(in EndpointInput, stored *Endpoint) error {
	name := strings.TrimSpace(in.HeaderName)
	if name == "" {
		return nil
	}
	if !headerName.MatchString(name) {
		return kernel.NewValidationError("headerName", "The header name may only contain letters, digits and - _ . characters.")
	}
	for _, f := range forbiddenHeaders {
		if strings.EqualFold(name, f) {
			return kernel.NewValidationError("headerName", "This header is set automatically and cannot be configured.")
		}
	}
	value := in.HeaderValue
	if value == "" && in.KeepHeader && stored != nil && strings.EqualFold(stored.HeaderName, name) {
		value = stored.HeaderValue
	}
	if value == "" {
		return kernel.NewValidationError("headerValue", "Enter the header value.")
	}
	if len(value) > maxHeaderValueLen || strings.ContainsAny(value, "\r\n\x00") {
		return kernel.NewValidationError("headerValue", "The header value is too long or contains line breaks.")
	}
	e.HeaderName, e.HeaderValue = name, value
	return nil
}

func finite(f *float64) bool { return f == nil || (!math.IsNaN(*f) && !math.IsInf(*f, 0)) }

// Status of a value against its thresholds.
type Status string

// Statuses; Unknown is used for values that are not numbers or not found.
const (
	StatusOK       Status = "ok"
	StatusWarning  Status = "warning"
	StatusCritical Status = "critical"
	StatusUnknown  Status = "unknown"
)

// Value is one extracted value.
type Value struct {
	Label string
	Unit  string
	// Text is the value for display; Number is set for numeric values
	// (also numeric strings) so the UI can format it.
	Text   string
	Number *float64
	Status Status
	// Error is set when the path did not match.
	Error string
}

// RawNumber is a JSON number with its exact digits (large integers stay exact).
type RawNumber string

// Composite is an object or array as compact JSON text.
type Composite string

// Evaluate turns a value found by the path into a Value with status. v is
// nil, bool, float64, string, RawNumber or Composite.
func (x Extraction) Evaluate(v any) Value {
	out := Value{Label: x.Label, Unit: x.Unit, Status: StatusUnknown}
	switch t := v.(type) {
	case nil:
		out.Text = "null"
	case bool:
		out.Text = strconv.FormatBool(t)
	case float64:
		out.Number = &t
	case RawNumber:
		if f, err := strconv.ParseFloat(string(t), 64); err == nil {
			out.Number = &f
		}
		out.Text = string(t)
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) {
			out.Number = &f
		}
		out.Text = truncate(t)
	case Composite:
		out.Text = truncate(string(t))
	default:
		out.Text = "?"
	}
	if out.Number != nil {
		if out.Text == "" {
			out.Text = strconv.FormatFloat(*out.Number, 'f', -1, 64)
		}
		out.Status = x.status(*out.Number)
	}
	return out
}

func (x Extraction) status(v float64) Status {
	reached := func(t *float64) bool {
		if t == nil {
			return false
		}
		if x.Below {
			return v <= *t
		}
		return v >= *t
	}
	switch {
	case reached(x.Crit):
		return StatusCritical
	case reached(x.Warn):
		return StatusWarning
	default:
		return StatusOK
	}
}

// Missing is the Value of an extraction whose path matched nothing.
func (x Extraction) Missing(reason string) Value {
	return Value{Label: x.Label, Unit: x.Unit, Status: StatusUnknown, Error: reason}
}

func truncate(s string) string {
	if utf8.RuneCountInString(s) <= maxTextRunes {
		return s
	}
	return string([]rune(s)[:maxTextRunes-1]) + "…"
}
