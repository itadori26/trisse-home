// Package domain holds the Feed and Item entities of the feed reader.
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

// Limits of the feed reader.
const (
	MaxItemsPerFeed   = 200
	MaxItemTitle      = 500
	MaxURLLength      = 2048
	maxFeedTitle      = 200
	maxErrorMessage   = 300
	untitledItemTitle = "(untitled)"
)

// Domain errors of the rss module.
var (
	ErrFeedNotFound      = kernel.NewPublicError(kernel.ErrNotFound, "The feed does not exist.")
	ErrItemNotFound      = kernel.NewPublicError(kernel.ErrNotFound, "The item does not exist.")
	ErrAlreadySubscribed = kernel.NewPublicError(kernel.ErrConflict, "You are already subscribed to this feed.")
)

// Feed is a subscribed RSS or Atom feed.
type Feed struct {
	ID            string
	URL           string
	Title         string
	SiteURL       string
	LastFetchedAt time.Time
	LastError     string
	CreatedAt     time.Time
}

// NewFeed validates the feed URL and creates a subscription.
func NewFeed(id, rawURL string, now time.Time) (Feed, error) {
	rawURL = strings.TrimSpace(rawURL)
	if !IsHTTPURL(rawURL) {
		return Feed{}, kernel.NewValidationError("url", "The feed URL must start with http:// or https://.")
	}
	return Feed{ID: id, URL: rawURL, Title: rawURL, CreatedAt: now}, nil
}

// RecordSuccess stores metadata of a successful fetch.
func (f *Feed) RecordSuccess(title, siteURL string, at time.Time) {
	if t := truncate(strings.TrimSpace(title), maxFeedTitle); t != "" {
		f.Title = t
	}
	if IsHTTPURL(siteURL) {
		f.SiteURL = siteURL
	}
	f.LastFetchedAt, f.LastError = at, ""
}

// RecordFailure stores the error of a failed fetch.
func (f *Feed) RecordFailure(err error, at time.Time) {
	f.LastFetchedAt, f.LastError = at, truncate(err.Error(), maxErrorMessage)
}

// Item is one entry of a feed.
type Item struct {
	ID          string
	FeedID      string
	GUID        string
	Title       string
	Link        string
	PublishedAt time.Time
	Read        bool
}

// NewItem sanitises an entry: only http(s) links are kept (no javascript:
// URLs from untrusted feeds), titles are trimmed, and a missing GUID falls
// back to the link or title.
func NewItem(id, feedID, guid, title, link string, published time.Time) Item {
	title = truncate(strings.Join(strings.Fields(title), " "), MaxItemTitle)
	if title == "" {
		title = untitledItemTitle
	}
	if !IsHTTPURL(link) {
		link = ""
	}
	guid = strings.TrimSpace(guid)
	if guid == "" {
		guid = link
	}
	if guid == "" {
		guid = title + "|" + published.UTC().Format(time.RFC3339)
	}
	return Item{ID: id, FeedID: feedID, GUID: truncate(guid, MaxURLLength), Title: title, Link: link, PublishedAt: published}
}

// IsHTTPURL reports whether raw is an absolute http(s) URL of sane length.
func IsHTTPURL(raw string) bool {
	if len(raw) > MaxURLLength {
		return false
	}
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func truncate(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes])
}
