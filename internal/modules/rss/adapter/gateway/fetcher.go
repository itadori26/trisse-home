package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"trisse-home/internal/modules/rss/usecase"
	"trisse-home/internal/shared/kernel"
)

// maxFeedBytes bounds the size of a downloaded feed.
const maxFeedBytes = 5 << 20

// HTTPFetcher implements usecase.FeedFetcher with the small parser of feedparse.go.
type HTTPFetcher struct{ client *http.Client }

// NewHTTPFetcher returns a fetcher using client.
func NewHTTPFetcher(client *http.Client) *HTTPFetcher { return &HTTPFetcher{client: client} }

// Fetch downloads and parses an RSS, Atom or JSON feed.
func (f *HTTPFetcher) Fetch(ctx context.Context, url string) (usecase.FetchedFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return usecase.FetchedFeed{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/feed+json, application/xml;q=0.9, */*;q=0.8")
	res, err := f.client.Do(req)
	if err != nil {
		return usecase.FetchedFeed{}, fmt.Errorf("download feed: %w: %w", kernel.ErrUnavailable, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return usecase.FetchedFeed{}, kernel.NewPublicErrorf(kernel.ErrUnavailable, "The server answered {status}.", "status", res.Status)
	}
	parsed, err := parseFeed(io.LimitReader(res.Body, maxFeedBytes))
	if err != nil {
		return usecase.FetchedFeed{}, kernel.NewPublicError(kernel.ErrInvalid, "this is not a valid RSS, Atom or JSON feed")
	}
	out := usecase.FetchedFeed{Title: parsed.Title, SiteURL: parsed.Link, Items: make([]usecase.FetchedItem, 0, len(parsed.Items))}
	for _, it := range parsed.Items {
		out.Items = append(out.Items, usecase.FetchedItem{GUID: it.ID, Title: it.Title, Link: it.Link, Published: it.Published})
	}
	return out, nil
}
