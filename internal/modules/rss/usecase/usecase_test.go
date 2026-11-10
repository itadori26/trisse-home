package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/rss/domain"
)

type memRepo struct {
	feeds map[string]domain.Feed
	items []domain.Item
}

func newMem() *memRepo { return &memRepo{feeds: map[string]domain.Feed{}} }

func (m *memRepo) ListFeeds(context.Context) ([]domain.Feed, error) {
	var out []domain.Feed
	for _, f := range m.feeds {
		out = append(out, f)
	}
	return out, nil
}
func (m *memRepo) GetFeed(_ context.Context, id string) (domain.Feed, error) {
	f, ok := m.feeds[id]
	if !ok {
		return f, domain.ErrFeedNotFound
	}
	return f, nil
}
func (m *memRepo) CreateFeed(_ context.Context, f domain.Feed) error {
	for _, e := range m.feeds {
		if e.URL == f.URL {
			return domain.ErrAlreadySubscribed
		}
	}
	m.feeds[f.ID] = f
	return nil
}
func (m *memRepo) UpdateFeed(_ context.Context, f domain.Feed) error { m.feeds[f.ID] = f; return nil }
func (m *memRepo) DeleteFeed(_ context.Context, id string) error {
	delete(m.feeds, id)
	m.items = slices.DeleteFunc(m.items, func(i domain.Item) bool { return i.FeedID == id })
	return nil
}
func (m *memRepo) AddItems(_ context.Context, items []domain.Item) (int, error) {
	n := 0
	for _, it := range items {
		if !slices.ContainsFunc(m.items, func(e domain.Item) bool { return e.FeedID == it.FeedID && e.GUID == it.GUID }) {
			m.items = append(m.items, it)
			n++
		}
	}
	return n, nil
}
func (m *memRepo) PruneItems(context.Context, string, int) error { return nil }
func (m *memRepo) ListItems(_ context.Context, f ItemFilter) ([]domain.Item, error) {
	var out []domain.Item
	for _, it := range m.items {
		if (f.FeedID == "" || it.FeedID == f.FeedID) && (!f.UnreadOnly || !it.Read) {
			out = append(out, it)
		}
	}
	return out, nil
}
func (m *memRepo) SetRead(_ context.Context, id string, read bool) error {
	for i := range m.items {
		if m.items[i].ID == id {
			m.items[i].Read = read
			return nil
		}
	}
	return domain.ErrItemNotFound
}
func (m *memRepo) MarkAllRead(_ context.Context, feedID string) error {
	for i := range m.items {
		if feedID == "" || m.items[i].FeedID == feedID {
			m.items[i].Read = true
		}
	}
	return nil
}
func (m *memRepo) UnreadCounts(context.Context) (map[string]int, error) {
	c := map[string]int{}
	for _, it := range m.items {
		if !it.Read {
			c[it.FeedID]++
		}
	}
	return c, nil
}

type fakeFetcher struct {
	feeds map[string]FetchedFeed
}

func (f *fakeFetcher) Fetch(_ context.Context, url string) (FetchedFeed, error) {
	ff, ok := f.feeds[url]
	if !ok {
		return FetchedFeed{}, errors.New("404 Not Found")
	}
	return ff, nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC) }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprint(i.n) }

func TestReader(t *testing.T) {
	ctx := context.Background()
	repo := newMem()
	fetcher := &fakeFetcher{feeds: map[string]FetchedFeed{
		"https://blog.example/feed": {Title: "Example", Items: []FetchedItem{{GUID: "1", Title: "One", Link: "https://blog.example/1"}, {GUID: "2", Title: "Two"}}},
	}}
	uc := NewReader(Deps{Repo: repo, Fetcher: fetcher, Clock: clk{}, IDs: &ids{}})

	if _, err := uc.Subscribe(ctx, "https://missing.example/feed"); err == nil {
		t.Fatal("expected error for unreachable feed")
	}
	feed, err := uc.Subscribe(ctx, "https://blog.example/feed")
	if err != nil || feed.Title != "Example" || feed.Unread != 2 {
		t.Fatalf("subscribe: %+v %v", feed, err)
	}
	if _, err := uc.Subscribe(ctx, "https://blog.example/feed"); !errors.Is(err, domain.ErrAlreadySubscribed) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	items, _ := uc.Items(ctx, ItemFilter{UnreadOnly: true})
	if len(items) != 2 || items[0].FeedTitle != "Example" {
		t.Fatalf("items: %+v", items)
	}
	if err := uc.MarkRead(ctx, items[0].ID, true); err != nil {
		t.Fatal(err)
	}
	if unread, _ := uc.Items(ctx, ItemFilter{UnreadOnly: true}); len(unread) != 1 {
		t.Fatal("mark read not applied")
	}
	// A refresh with the same items must not duplicate them or reset read state.
	if err := uc.RefreshAll(ctx); err != nil {
		t.Fatal(err)
	}
	if all, _ := uc.Items(ctx, ItemFilter{}); len(all) != 2 {
		t.Fatalf("items duplicated: %d", len(all))
	}
	if err := uc.MarkAllRead(ctx, feed.ID); err != nil {
		t.Fatal(err)
	}
	feeds, _ := uc.Feeds(ctx)
	if feeds[0].Unread != 0 {
		t.Fatal("mark all read failed")
	}
	delete(fetcher.feeds, "https://blog.example/feed")
	if err := uc.RefreshAll(ctx); err == nil {
		t.Fatal("expected refresh error")
	}
	if f, _ := repo.GetFeed(ctx, feed.ID); f.LastError == "" {
		t.Fatal("failure not recorded on feed")
	}
	if err := uc.Unsubscribe(ctx, feed.ID); err != nil || len(repo.items) != 0 {
		t.Fatalf("unsubscribe: %v", err)
	}
}
