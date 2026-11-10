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
	dto := FeedDTO{ID: v.ID, URL: v.URL, Title: v.Title, SiteURL: v.SiteURL, LastError: v.LastError, Unread: v.Unread}
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
