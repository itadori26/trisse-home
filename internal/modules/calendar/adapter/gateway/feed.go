package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/ical"
)

// maxFeedBytes bounds one ICS download; timetables with years of history
// can be a few megabytes.
const maxFeedBytes = 8 << 20

// Feed implements usecase.FeedSource for ICS subscriptions (#98). It keeps
// the parsed feed in memory and downloads it again only after the
// subscription's refresh interval, with ETag/Last-Modified so an unchanged
// feed costs a 304. Recurrences are expanded per requested range.
type Feed struct {
	client *http.Client
	loc    *time.Location
	now    func() time.Time
	mu     sync.Mutex
	cache  map[string]*feedEntry
}

type feedEntry struct {
	events       []ical.Event
	fetchedAt    time.Time
	etag         string
	lastModified string
}

// NewFeed returns a feed client; floating times are interpreted in loc.
func NewFeed(client *http.Client, loc *time.Location, now func() time.Time) *Feed {
	return &Feed{client: client, loc: loc, now: now, cache: map[string]*feedEntry{}}
}

// Events implements usecase.FeedSource.
func (f *Feed) Events(ctx context.Context, s domain.Subscription, from, to time.Time) ([]domain.Event, error) {
	entry, err := f.load(ctx, s)
	if err != nil {
		return nil, err
	}
	var out []domain.Event
	for _, e := range entry.events {
		for _, occ := range ical.Expand(e, from, to) {
			out = append(out, domain.Event{UID: occ.UID, Summary: occ.Summary, Location: occ.Location,
				Start: occ.Start, End: occ.End, AllDay: occ.AllDay})
		}
	}
	return out, nil
}

// load returns the cached feed or downloads it when the interval passed.
// A failed refresh keeps serving the last good copy.
func (f *Feed) load(ctx context.Context, s domain.Subscription) (*feedEntry, error) {
	key := s.ID + "\x00" + s.URL
	f.mu.Lock()
	defer f.mu.Unlock()
	old := f.cache[key]
	if old != nil && f.now().Sub(old.fetchedAt) < s.RefreshInterval {
		return old, nil
	}
	fresh, err := f.download(ctx, s.URL, old)
	if err != nil {
		if old != nil {
			return old, nil
		}
		return nil, err
	}
	f.cache[key] = fresh
	return fresh, nil
}

func (f *Feed) download(ctx context.Context, url string, old *feedEntry) (*feedEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "text/calendar, */*;q=0.5")
	if old != nil {
		if old.etag != "" {
			req.Header.Set("If-None-Match", old.etag)
		}
		if old.lastModified != "" {
			req.Header.Set("If-Modified-Since", old.lastModified)
		}
	}
	res, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusNotModified && old != nil:
		next := *old
		next.fetchedAt = f.now()
		return &next, nil
	case res.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("server answered %s", res.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, maxFeedBytes))
	if err != nil {
		return nil, fmt.Errorf("read calendar: %w", err)
	}
	events, err := ical.Parse(string(raw), f.loc)
	if err != nil {
		return nil, err
	}
	return &feedEntry{events: events, fetchedAt: f.now(), etag: res.Header.Get("ETag"), lastModified: res.Header.Get("Last-Modified")}, nil
}
