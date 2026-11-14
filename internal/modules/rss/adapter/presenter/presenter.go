package presenter

import (
	"time"

	"trisse-home/internal/modules/rss/usecase"
)

// FeedDTO is a subscribed feed.
type FeedDTO struct {
	ID            string     `json:"id"`
	URL           string     `json:"url"`
	Title         string     `json:"title"`
	SiteURL       string     `json:"siteUrl"`
	LastError     string     `json:"lastError"`
	LastFetchedAt *time.Time `json:"lastFetchedAt"`
	Unread        int        `json:"unread"`
	// Category groups feeds ("" = none; OPML folders, #97).
	Category string `json:"category"`
}

// InvalidEntryDTO is an OPML entry that was not imported.
type InvalidEntryDTO struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Reason string `json:"reason"`
}

// ImportReportDTO is returned by POST /api/rss/opml.
type ImportReportDTO struct {
	Added      int               `json:"added"`
	Duplicates int               `json:"duplicates"`
	Invalid    []InvalidEntryDTO `json:"invalid"`
}

// ImportReport maps an import report.
func ImportReport(r usecase.ImportReport) ImportReportDTO {
	dto := ImportReportDTO{Added: r.Added, Duplicates: r.Duplicates, Invalid: make([]InvalidEntryDTO, 0, len(r.Invalid))}
	for _, i := range r.Invalid {
		dto.Invalid = append(dto.Invalid, InvalidEntryDTO(i))
	}
	return dto
}

// ItemDTO is a feed entry.
type ItemDTO struct {
	ID          string    `json:"id"`
	FeedID      string    `json:"feedId"`
	FeedTitle   string    `json:"feedTitle"`
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	PublishedAt time.Time `json:"publishedAt"`
	Read        bool      `json:"read"`
}

// Feed maps a FeedView.
func Feed(v usecase.FeedView) FeedDTO {
	dto := FeedDTO{ID: v.ID, URL: v.URL, Title: v.Title, SiteURL: v.SiteURL, LastError: v.LastError, Unread: v.Unread, Category: v.Category}
	if !v.LastFetchedAt.IsZero() {
		t := v.LastFetchedAt
		dto.LastFetchedAt = &t
	}
	return dto
}

// Feeds maps feed views.
func Feeds(vs []usecase.FeedView) []FeedDTO {
	out := make([]FeedDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, Feed(v))
	}
	return out
}

// Items maps item views.
func Items(vs []usecase.ItemView) []ItemDTO {
	out := make([]ItemDTO, 0, len(vs))
	for _, v := range vs {
		out = append(out, ItemDTO(v))
	}
	return out
}
