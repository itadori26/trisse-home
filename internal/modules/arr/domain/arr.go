package domain

import (
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// App is a *arr application (#86).
type App string

// Supported apps; Sonarr and Radarr share the v3 API, Prowlarr uses v1.
const (
	AppSonarr   App = "sonarr"
	AppRadarr   App = "radarr"
	AppProwlarr App = "prowlarr"
)

// Limits of the module.
const (
	MaxInstances = 10
	maxName      = 40
	// UpcomingDays is how far ahead releases are listed.
	UpcomingDays = 14
	// MaxQueue is how many downloads are listed per instance.
	MaxQueue = 20
)

// ErrNotConfigured is returned while no instance is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "No Sonarr, Radarr or Prowlarr instance is configured yet.")

// ErrInstanceNotFound is returned for an unknown instance ID.
var ErrInstanceNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The instance does not exist.")

// ErrTooManyInstances is returned when the limit is reached.
var ErrTooManyInstances = kernel.NewPublicErrorf(kernel.ErrInvalid, "At most {max} instances are supported.", "max", MaxInstances)

// Instance is one configured app.
type Instance struct {
	ID     string
	App    App
	Name   string
	URL    string
	APIKey string
}

// NewInstance validates an instance; the name defaults to the app.
func NewInstance(id, app, name, rawURL, apiKey string) (Instance, error) {
	a := App(strings.TrimSpace(app))
	if !slices.Contains([]App{AppSonarr, AppRadarr, AppProwlarr}, a) {
		return Instance{}, kernel.NewValidationError("app", "Choose Sonarr, Radarr or Prowlarr.")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = a.Title()
	}
	if utf8.RuneCountInString(name) > maxName {
		return Instance{}, kernel.NewValidationError("name", "The name must be at most 40 characters long.")
	}
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Instance{}, kernel.NewValidationError("url", "The URL must look like http://192.168.1.2:8989.")
	}
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return Instance{}, kernel.NewValidationError("apiKey", "The API key is required (Settings → General).")
	}
	return Instance{ID: id, App: a, Name: name, URL: rawURL, APIKey: apiKey}, nil
}

// Title is the product name of the app.
func (a App) Title() string {
	switch a {
	case AppSonarr:
		return "Sonarr"
	case AppRadarr:
		return "Radarr"
	case AppProwlarr:
		return "Prowlarr"
	}
	return string(a)
}

// HasLibrary reports whether the app has releases, a queue and missing
// items; Prowlarr only manages indexers and reports health.
func (a App) HasLibrary() bool { return a == AppSonarr || a == AppRadarr }

// Release is an upcoming episode or movie.
type Release struct {
	Instance string
	Title    string
	// Detail is e.g. "S02E05 · Episode title" or the movie year.
	Detail  string
	Date    time.Time
	HasFile bool
}

// QueueItem is a running or stuck download.
type QueueItem struct {
	Instance string
	Title    string
	Status   string
	// Progress is between 0 and 100.
	Progress float64
	TimeLeft time.Duration
	// Error is set when the download needs attention.
	Error string
}

// Severity of a health issue.
type Severity string

// Health severities as reported by the apps.
const (
	SeverityNotice  Severity = "notice"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// HealthIssue is a problem reported by an app, e.g. an unreachable indexer.
type HealthIssue struct {
	Instance string
	Severity Severity
	Message  string
}

// Snapshot is what one instance reports.
type Snapshot struct {
	Upcoming []Release
	Queue    []QueueItem
	Health   []HealthIssue
	Missing  int
}

// Overview merges the snapshots of all instances.
type Overview struct {
	Upcoming []Release
	Queue    []QueueItem
	Health   []HealthIssue
	// Missing counts monitored items without a file per instance name.
	Missing map[string]int
	// Errors lists instances that could not be reached.
	Errors []string
}

// Merge combines snapshots in instance order and sorts releases by date.
func Merge(names []string, snaps []Snapshot) Overview {
	o := Overview{Missing: map[string]int{}}
	for i, s := range snaps {
		o.Upcoming = append(o.Upcoming, s.Upcoming...)
		o.Queue = append(o.Queue, s.Queue...)
		o.Health = append(o.Health, s.Health...)
		if s.Missing > 0 {
			o.Missing[names[i]] = s.Missing
		}
	}
	slices.SortStableFunc(o.Upcoming, func(a, b Release) int { return a.Date.Compare(b.Date) })
	slices.SortStableFunc(o.Health, func(a, b HealthIssue) int { return severityRank(b.Severity) - severityRank(a.Severity) })
	return o
}

func severityRank(s Severity) int {
	switch s {
	case SeverityError:
		return 2
	case SeverityWarning:
		return 1
	}
	return 0
}
