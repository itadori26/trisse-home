package domain

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

// OPML import and export of subscriptions (#97).

// Limits of an import.
const (
	// MaxImportFeeds bounds one import.
	MaxImportFeeds = 500
	// MaxImportBytes bounds the OPML file.
	MaxImportBytes = 2 << 20
	maxCategory    = 60
)

// ImportEntry is one feed of an OPML file.
type ImportEntry struct {
	Title    string
	URL      string
	SiteURL  string
	Category string
}

// InvalidEntry is an entry that was not imported, with the reason.
type InvalidEntry struct {
	Title, URL, Reason string
}

// ImportPlan says what an import adds and what it skips.
type ImportPlan struct {
	Add        []ImportEntry
	Duplicates int
	Invalid    []InvalidEntry
}

// CleanCategory trims a category and limits its length.
func CleanCategory(c string) string {
	return truncate(strings.Join(strings.Fields(c), " "), maxCategory)
}

// sameFeed is the duplicate rule: URLs that differ only in case of scheme
// and host, a trailing slash or a fragment are the same feed.
func sameFeed(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(raw)
	}
	u.Scheme, u.Host, u.Fragment = strings.ToLower(u.Scheme), strings.ToLower(u.Host), ""
	return strings.TrimSuffix(u.String(), "/")
}

// PlanImport decides which entries an import adds: valid http(s) URLs that
// are neither subscribed already nor repeated in the file, at most
// MaxImportFeeds; everything else is counted or reported.
func PlanImport(existing []Feed, entries []ImportEntry) ImportPlan {
	seen := make(map[string]bool, len(existing)+len(entries))
	for _, f := range existing {
		seen[sameFeed(f.URL)] = true
	}
	var plan ImportPlan
	for _, e := range entries {
		e.URL, e.Title, e.Category = strings.TrimSpace(e.URL), strings.TrimSpace(e.Title), CleanCategory(e.Category)
		switch {
		case !IsHTTPURL(e.URL):
			plan.Invalid = append(plan.Invalid, InvalidEntry{Title: e.Title, URL: truncate(e.URL, 200), Reason: "not an http(s) feed address"})
		case seen[sameFeed(e.URL)]:
			plan.Duplicates++
		case len(plan.Add) >= MaxImportFeeds:
			plan.Invalid = append(plan.Invalid, InvalidEntry{Title: e.Title, URL: e.URL, Reason: "more than 500 feeds in one import"})
		default:
			seen[sameFeed(e.URL)] = true
			if !IsHTTPURL(e.SiteURL) {
				e.SiteURL = ""
			}
			if utf8.RuneCountInString(e.Title) > maxFeedTitle {
				e.Title = truncate(e.Title, maxFeedTitle)
			}
			plan.Add = append(plan.Add, e)
		}
	}
	return plan
}
