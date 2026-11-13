package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"trisse-home/internal/modules/releases/domain"
)

// Public endpoints of MusicBrainz and the Cover Art Archive.
const (
	DefaultMusicBrainzURL = "https://musicbrainz.org/ws/2"
	DefaultCoverArtURL    = "https://coverartarchive.org"
)

const (
	maxJSONBytes  = 2 << 20
	maxCoverBytes = 1 << 20
	// minInterval is MusicBrainz's rate limit: one request per second.
	minInterval = 1100 * time.Millisecond
)

// MusicBrainz implements usecase.ReleaseProvider. All MusicBrainz requests
// share one limiter, so search and refresh never exceed one per second;
// the shared HTTP client sends the required User-Agent.
type MusicBrainz struct {
	client   *http.Client
	baseURL  string
	coverURL string
	// wait sleeps for d unless ctx ends first (replaced in tests).
	wait func(ctx context.Context, d time.Duration) error
	now  func() time.Time
	mu   sync.Mutex
	last time.Time
}

// NewMusicBrainz returns a client for the given base URLs.
func NewMusicBrainz(client *http.Client, baseURL, coverURL string) *MusicBrainz {
	return &MusicBrainz{client: client, baseURL: strings.TrimRight(baseURL, "/"), coverURL: strings.TrimRight(coverURL, "/"),
		wait: sleep, now: time.Now}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// SearchArtists implements usecase.ReleaseProvider.
func (m *MusicBrainz) SearchArtists(ctx context.Context, query string) ([]domain.Artist, error) {
	var p struct {
		Artists []struct {
			ID             string `json:"id"`
			Name           string `json:"name"`
			Disambiguation string `json:"disambiguation"`
		} `json:"artists"`
	}
	q := url.Values{"query": {query}, "limit": {"10"}, "fmt": {"json"}}
	if err := m.get(ctx, "/artist?"+q.Encode(), &p); err != nil {
		return nil, err
	}
	out := make([]domain.Artist, 0, len(p.Artists))
	for _, a := range p.Artists {
		out = append(out, domain.Artist{ID: a.ID, Name: a.Name, Disambiguation: a.Disambiguation})
	}
	return out, nil
}

// ReleasesSince implements usecase.ReleaseProvider with one Lucene search
// per artist (arid plus a date range), instead of paging the discography.
func (m *MusicBrainz) ReleasesSince(ctx context.Context, a domain.Artist, since time.Time) ([]domain.Release, error) {
	var p struct {
		Groups []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			PrimaryType string `json:"primary-type"`
			FirstDate   string `json:"first-release-date"`
		} `json:"release-groups"`
	}
	query := fmt.Sprintf("arid:%s AND firstreleasedate:[%s TO *]", a.ID, since.Format(time.DateOnly))
	q := url.Values{"query": {query}, "limit": {"100"}, "fmt": {"json"}}
	if err := m.get(ctx, "/release-group?"+q.Encode(), &p); err != nil {
		return nil, err
	}
	out := make([]domain.Release, 0, len(p.Groups))
	for _, g := range p.Groups {
		date, ok := domain.ParseDate(g.FirstDate)
		if !ok || date.Before(since) {
			continue
		}
		kind := g.PrimaryType
		if kind == "" {
			kind = "Other"
		}
		out = append(out, domain.Release{ID: g.ID, Title: g.Title, Artist: a.Name, Type: kind, Date: date})
	}
	return out, nil
}

// Cover implements usecase.ReleaseProvider: the 250 px front cover of a
// release group (the archive redirects to the image host).
func (m *MusicBrainz) Cover(ctx context.Context, releaseID string) (domain.Cover, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.coverURL+"/release-group/"+url.PathEscape(releaseID)+"/front-250", nil)
	if err != nil {
		return domain.Cover{}, fmt.Errorf("build request: %w", err)
	}
	res, err := m.client.Do(req)
	if err != nil {
		return domain.Cover{}, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return domain.Cover{}, fmt.Errorf("no cover: %s", res.Status)
	}
	ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if !strings.HasPrefix(ct, "image/") || ct == "image/svg+xml" {
		return domain.Cover{}, fmt.Errorf("not a raster image: %q", ct)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxCoverBytes+1))
	if err != nil || len(data) > maxCoverBytes {
		return domain.Cover{}, fmt.Errorf("cover too large or unreadable")
	}
	return domain.Cover{ContentType: ct, Data: data}, nil
}

// get waits for the rate limit and decodes a MusicBrainz JSON answer.
func (m *MusicBrainz) get(ctx context.Context, path string, dst any) error {
	if err := m.throttle(ctx); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	res, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected answer from MusicBrainz: %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxJSONBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from MusicBrainz: %w", err)
	}
	return nil
}

// throttle keeps at least minInterval between two MusicBrainz requests.
func (m *MusicBrainz) throttle(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if wait := minInterval - m.now().Sub(m.last); m.last != (time.Time{}) && wait > 0 {
		if err := m.wait(ctx, wait); err != nil {
			return err
		}
	}
	m.last = m.now()
	return nil
}
