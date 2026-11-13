package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/arr/domain"
)

// Library implements usecase.Provider for Sonarr and Radarr (API v3):
// calendar, queue, health and the number of missing items.
type Library struct {
	c   client
	app domain.App
}

// NewSonarr returns the Sonarr adapter.
func NewSonarr(h *http.Client) *Library {
	return &Library{c: client{http: h, version: "v3"}, app: domain.AppSonarr}
}

// NewRadarr returns the Radarr adapter.
func NewRadarr(h *http.Client) *Library {
	return &Library{c: client{http: h, version: "v3"}, app: domain.AppRadarr}
}

// Check implements usecase.Provider.
func (l *Library) Check(ctx context.Context, in domain.Instance) error { return l.c.check(ctx, in) }

type calendarEntry struct {
	// Sonarr
	Title         string    `json:"title"`
	SeasonNumber  int       `json:"seasonNumber"`
	EpisodeNumber int       `json:"episodeNumber"`
	AirDateUtc    time.Time `json:"airDateUtc"`
	Series        struct {
		Title string `json:"title"`
	} `json:"series"`
	// Radarr
	Year            int        `json:"year"`
	InCinemas       *time.Time `json:"inCinemas"`
	DigitalRelease  *time.Time `json:"digitalRelease"`
	PhysicalRelease *time.Time `json:"physicalRelease"`
	HasFile         bool       `json:"hasFile"`
}

type queuePayload struct {
	Records []struct {
		Title                 string  `json:"title"`
		Status                string  `json:"status"`
		TrackedDownloadStatus string  `json:"trackedDownloadStatus"`
		Size                  float64 `json:"size"`
		SizeLeft              float64 `json:"sizeleft"`
		TimeLeft              string  `json:"timeleft"`
		ErrorMessage          string  `json:"errorMessage"`
		StatusMessages        []struct {
			Messages []string `json:"messages"`
		} `json:"statusMessages"`
	} `json:"records"`
}

// Snapshot implements usecase.Provider.
func (l *Library) Snapshot(ctx context.Context, in domain.Instance, from, to time.Time) (domain.Snapshot, error) {
	var s domain.Snapshot
	q := url.Values{"start": {from.UTC().Format(time.RFC3339)}, "end": {to.UTC().Format(time.RFC3339)}}
	if l.app == domain.AppSonarr {
		q.Set("includeSeries", "true")
	}
	var cal []calendarEntry
	if err := l.c.get(ctx, in, "/calendar?"+q.Encode(), &cal); err != nil {
		return s, err
	}
	for _, e := range cal {
		if r, ok := l.release(in.Name, e, from, to); ok {
			s.Upcoming = append(s.Upcoming, r)
		}
	}
	var queue queuePayload
	if err := l.c.get(ctx, in, "/queue?pageSize="+strconv.Itoa(domain.MaxQueue), &queue); err != nil {
		return s, err
	}
	for _, r := range queue.Records {
		item := domain.QueueItem{Instance: in.Name, Title: r.Title, Status: strings.ToLower(r.Status), TimeLeft: parseTimeLeft(r.TimeLeft)}
		if r.Size > 0 {
			item.Progress = (r.Size - r.SizeLeft) / r.Size * 100
		}
		if r.TrackedDownloadStatus == "warning" || r.TrackedDownloadStatus == "error" || r.ErrorMessage != "" {
			item.Error = r.ErrorMessage
			if item.Error == "" && len(r.StatusMessages) > 0 && len(r.StatusMessages[0].Messages) > 0 {
				item.Error = r.StatusMessages[0].Messages[0]
			}
		}
		s.Queue = append(s.Queue, item)
	}
	var err error
	if s.Health, err = l.c.health(ctx, in); err != nil {
		return s, err
	}
	var missing struct {
		TotalRecords int `json:"totalRecords"`
	}
	if err := l.c.get(ctx, in, "/wanted/missing?pageSize=1", &missing); err != nil {
		return s, err
	}
	s.Missing = missing.TotalRecords
	return s, nil
}

// release maps a calendar entry; Radarr movies use the first release date
// within the range (cinema, digital or physical).
func (l *Library) release(instance string, e calendarEntry, from, to time.Time) (domain.Release, bool) {
	if l.app == domain.AppSonarr {
		return domain.Release{Instance: instance, Title: e.Series.Title, HasFile: e.HasFile, Date: e.AirDateUtc,
			Detail: fmt.Sprintf("S%02dE%02d · %s", e.SeasonNumber, e.EpisodeNumber, e.Title)}, !e.AirDateUtc.IsZero()
	}
	for _, d := range []*time.Time{e.InCinemas, e.DigitalRelease, e.PhysicalRelease} {
		if d != nil && !d.Before(from) && d.Before(to) {
			return domain.Release{Instance: instance, Title: e.Title, Detail: strconv.Itoa(e.Year), Date: *d, HasFile: e.HasFile}, true
		}
	}
	return domain.Release{}, false
}

// parseTimeLeft reads the "[d.]hh:mm:ss" format of the *arr APIs.
func parseTimeLeft(s string) time.Duration {
	var days int
	if d, rest, ok := strings.Cut(s, "."); ok && !strings.Contains(rest, ".") && strings.Count(rest, ":") == 2 {
		days, _ = strconv.Atoi(d)
		s = rest
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	sec, _ := strconv.ParseFloat(parts[2], 64)
	return time.Duration(days)*24*time.Hour + time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec*float64(time.Second))
}

// Prowlarr implements usecase.Provider for Prowlarr (API v1), which only
// reports health (e.g. failing indexers).
type Prowlarr struct{ c client }

// NewProwlarr returns the Prowlarr adapter.
func NewProwlarr(h *http.Client) *Prowlarr { return &Prowlarr{c: client{http: h, version: "v1"}} }

// Check implements usecase.Provider.
func (p *Prowlarr) Check(ctx context.Context, in domain.Instance) error { return p.c.check(ctx, in) }

// Snapshot implements usecase.Provider.
func (p *Prowlarr) Snapshot(ctx context.Context, in domain.Instance, _, _ time.Time) (domain.Snapshot, error) {
	health, err := p.c.health(ctx, in)
	return domain.Snapshot{Health: health}, err
}
