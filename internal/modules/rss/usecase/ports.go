package usecase

import (
	"context"
	"time"

	"trisse-home/internal/modules/rss/domain"
	"trisse-home/internal/shared/kernel"
)

// FetchedItem is an entry as delivered by the FeedFetcher.
type FetchedItem struct {
	GUID      string
	Title     string
	Link      string
	Published time.Time
}

// FetchedFeed is a parsed feed as delivered by the FeedFetcher.
type FetchedFeed struct {
	Title   string
	SiteURL string
	Items   []FetchedItem
}

// FeedFetcher downloads and parses a feed (output port).
type FeedFetcher interface {
	Fetch(ctx context.Context, url string) (FetchedFeed, error)
}

// ItemFilter narrows the item list.
type ItemFilter struct {
	FeedID     string
	UnreadOnly bool
	Limit      int
}

// Repository persists feeds and items (output port).
type Repository interface {
	ListFeeds(ctx context.Context) ([]domain.Feed, error)
	GetFeed(ctx context.Context, id string) (domain.Feed, error)
	// CreateFeed returns domain.ErrAlreadySubscribed for duplicate URLs.
	CreateFeed(ctx context.Context, f domain.Feed) error
	UpdateFeed(ctx context.Context, f domain.Feed) error
	// DeleteFeed removes the feed and all its items.
	DeleteFeed(ctx context.Context, id string) error
	// AddItems inserts items whose GUID is new for the feed and keeps the read
	// state of existing ones; it returns the number of inserted items.
	AddItems(ctx context.Context, items []domain.Item) (int, error)
	// PruneItems keeps only the newest `keep` items of a feed.
	PruneItems(ctx context.Context, feedID string, keep int) error
	ListItems(ctx context.Context, f ItemFilter) ([]domain.Item, error)
	SetRead(ctx context.Context, itemID string, read bool) error
	// MarkAllRead marks all items (of one feed if feedID != "") as read.
	MarkAllRead(ctx context.Context, feedID string) error
	UnreadCounts(ctx context.Context) (map[string]int, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo    Repository
	Fetcher FeedFetcher
	Clock   kernel.Clock
	IDs     kernel.IDGenerator
}
