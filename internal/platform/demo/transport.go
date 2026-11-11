package demo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Transport answers outbound requests with fixtures and refuses everything
// else, so a demo instance never contacts third parties.
type Transport struct {
	// Now returns the current time; fixtures are generated relative to it.
	Now func() time.Time
}

// RoundTrip implements http.RoundTripper.
func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		_ = req.Body.Close()
	}
	now := time.Now()
	if t.Now != nil {
		now = t.Now()
	}
	host, path := req.URL.Hostname(), req.URL.Path
	switch {
	case host == "api.open-meteo.com":
		return respond(req, http.StatusOK, "application/json", forecast(now))
	case host == "geocoding-api.open-meteo.com":
		return respond(req, http.StatusOK, "application/json", geocoding)
	case strings.HasSuffix(path, "/containers/json"):
		return respond(req, http.StatusOK, "application/json", containers(now))
	case strings.HasSuffix(host, "go.dev") || strings.HasSuffix(host, "hnrss.org") || strings.Contains(path, "feed") || strings.HasSuffix(path, ".xml"):
		return respond(req, http.StatusOK, "application/atom+xml", feed(req.URL.Host, now))
	}
	return respond(req, http.StatusServiceUnavailable, "text/plain", "outbound requests are disabled in demo mode")
}

func respond(req *http.Request, status int, contentType, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: status, Status: fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1,
		Header:        http.Header{"Content-Type": {contentType}},
		Body:          io.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
	}, nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v) // only maps, slices and numbers: cannot fail
	return string(b)
}

// forecast is a plausible seven-day forecast starting today.
func forecast(now time.Time) string {
	days := 7
	times := make([]string, days)
	for i := range days {
		times[i] = now.AddDate(0, 0, i).Format(time.DateOnly)
	}
	precip := []any{10, 40, 80, 20, 0, 5, 60}
	return mustJSON(map[string]any{
		"timezone": "Europe/Berlin",
		"current": map[string]any{
			"temperature_2m": 14.2, "apparent_temperature": 12.8, "relative_humidity_2m": 68,
			"wind_speed_10m": 11.5, "weather_code": 2, "is_day": 1,
		},
		"daily": map[string]any{
			"time":                          times,
			"weather_code":                  []int{2, 3, 61, 1, 0, 2, 63},
			"temperature_2m_max":            []float64{16, 15, 12, 14, 17, 18, 13},
			"temperature_2m_min":            []float64{8, 9, 7, 6, 7, 9, 8},
			"precipitation_probability_max": precip,
		},
	})
}

const geocoding = `{"results":[{"name":"Berlin","admin1":"Berlin","country":"Germany","latitude":52.52,"longitude":13.41}]}`

// containers is a made-up container list for the Docker widget.
func containers(now time.Time) string {
	type c struct {
		ID      string            `json:"Id"`
		Names   []string          `json:"Names"`
		Image   string            `json:"Image"`
		State   string            `json:"State"`
		Status  string            `json:"Status"`
		Created int64             `json:"Created"`
		Labels  map[string]string `json:"Labels"`
	}
	created := now.Add(-72 * time.Hour).Unix()
	list := []c{
		{"a1", []string{"/trisse-home"}, "ghcr.io/itadori26/trisse-home:latest", "running", "Up 3 days (healthy)", created, nil},
		{"b2", []string{"/jellyfin"}, "jellyfin/jellyfin:latest", "running", "Up 3 days", created, nil},
		{"c3", []string{"/adguardhome"}, "adguard/adguardhome:latest", "running", "Up 3 days", created, nil},
		{"d4", []string{"/paperless"}, "ghcr.io/paperless-ngx/paperless-ngx:latest", "running", "Up 2 hours", created, nil},
		{"e5", []string{"/backup-job"}, "restic/restic:latest", "exited", "Exited (0) 5 hours ago", created, nil},
	}
	return mustJSON(list)
}

// feed is a small Atom feed with entries from the last days.
func feed(host string, now time.Time) string {
	titles := []string{
		"Welcome to the trisse-home demo",
		"Widgets can be moved and resized with the mouse or keyboard",
		"Press / or Ctrl+K to search bookmarks, services, notes and to-dos",
		"Everything runs on your own server: no account, no tracking",
	}
	if !strings.HasSuffix(host, "go.dev") {
		titles = []string{
			"Show HN: A self-hosted start page for the home server",
			"Ask HN: How do you monitor your homelab?",
			"Wake-on-LAN from a Docker container, explained",
			"Why I moved my calendar to CalDAV",
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="utf-8"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Demo feed (%s)</title><id>urn:demo:%s</id><updated>%s</updated>`,
		host, host, now.Format(time.RFC3339))
	for i, title := range titles {
		ts := now.Add(-time.Duration(i*9+1) * time.Hour).Format(time.RFC3339)
		fmt.Fprintf(&b, `<entry><title>%s</title><id>urn:demo:%s:%d</id><link href="https://github.com/itadori26/trisse-home"/><updated>%s</updated><published>%s</published></entry>`,
			title, host, i, ts, ts)
	}
	b.WriteString(`</feed>`)
	return b.String()
}

// HealthyTransport answers every request with 200 OK, so service tiles of
// the demo show as reachable without probing anything.
type HealthyTransport struct{}

// RoundTrip implements http.RoundTripper.
func (HealthyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		_ = req.Body.Close()
	}
	return respond(req, http.StatusOK, "text/plain", "ok")
}
