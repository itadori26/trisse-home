package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/modules/rss/domain"
	"trisse-home/internal/shared/kernel"
)

// Item list limits.
const (
	DefaultItemLimit = 50
	MaxItemLimit     = 200
)

// FeedView is a feed with its unread count.
type FeedView struct {
	ID, URL, Title, SiteURL, LastError string
	Category                           string
	LastFetchedAt                      time.Time
	Unread                             int
}

// ItemView is an item together with its feed title.
type ItemView struct {
	ID, FeedID, FeedTitle, Title, Link string
	PublishedAt                        time.Time
	Read                               bool
}

// Reader is the input port of the feed reader. Each method is one use case.
type Reader interface {
	Subscribe(ctx context.Context, url string) (FeedView, error)
	Feeds(ctx context.Context) ([]FeedView, error)
	Unsubscribe(ctx context.Context, feedID string) error
	Items(ctx context.Context, f ItemFilter) ([]ItemView, error)
	MarkRead(ctx context.Context, itemID string, read bool) error
	MarkAllRead(ctx context.Context, feedID string) error
	RefreshAll(ctx context.Context) error
	// ImportOPML subscribes to the feeds of an OPML file (#97). Feeds are not
	// fetched here; the next refresh loads them.
	ImportOPML(ctx context.Context, data []byte) (ImportReport, error)
	// ExportOPML writes all subscriptions as OPML 2.0.
	ExportOPML(ctx context.Context) ([]byte, error)
}

// ImportReport says what an OPML import did.
type ImportReport struct {
	Added      int
	Duplicates int
	Invalid    []domain.InvalidEntry
}

type reader struct{ d Deps }

// NewReader returns the feed reader use cases.
func NewReader(d Deps) Reader { return &reader{d: d} }

// Subscribe implements Reader. The feed is fetched once immediately so that
// invalid URLs are reported right away instead of failing silently later.
func (uc *reader) Subscribe(ctx context.Context, url string) (FeedView, error) {
	f, err := domain.NewFeed(uc.d.IDs.NewID(), url, uc.d.Clock.Now())
	if err != nil {
		return FeedView{}, err
	}
	fetched, err := uc.d.Fetcher.Fetch(ctx, f.URL)
	if err != nil {
		return FeedView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "The feed could not be loaded: {detail}", "detail", publicReason(err))
	}
	if err := uc.d.Repo.CreateFeed(ctx, f); err != nil {
		return FeedView{}, err
	}
	if err := uc.store(ctx, &f, fetched); err != nil {
		return FeedView{}, err
	}
	return uc.view(ctx, f)
}

// Feeds implements Reader.
func (uc *reader) Feeds(ctx context.Context) ([]FeedView, error) {
	feeds, err := uc.d.Repo.ListFeeds(ctx)
	if err != nil {
		return nil, fmt.Errorf("list feeds: %w", err)
	}
	counts, err := uc.d.Repo.UnreadCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("count unread: %w", err)
	}
	out := make([]FeedView, 0, len(feeds))
	for _, f := range feeds {
		out = append(out, toFeedView(f, counts[f.ID]))
	}
	return out, nil
}

// Unsubscribe implements Reader.
func (uc *reader) Unsubscribe(ctx context.Context, feedID string) error {
	if _, err := uc.d.Repo.GetFeed(ctx, feedID); err != nil {
		return err
	}
	return uc.d.Repo.DeleteFeed(ctx, feedID)
}

// Items implements Reader.
func (uc *reader) Items(ctx context.Context, f ItemFilter) ([]ItemView, error) {
	if f.Limit <= 0 {
		f.Limit = DefaultItemLimit
	}
	f.Limit = min(f.Limit, MaxItemLimit)
	feeds, err := uc.d.Repo.ListFeeds(ctx)
	if err != nil {
		return nil, fmt.Errorf("list feeds: %w", err)
	}
	titles := make(map[string]string, len(feeds))
	for _, fd := range feeds {
		titles[fd.ID] = fd.Title
	}
	items, err := uc.d.Repo.ListItems(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	out := make([]ItemView, 0, len(items))
	for _, it := range items {
		out = append(out, ItemView{ID: it.ID, FeedID: it.FeedID, FeedTitle: titles[it.FeedID], Title: it.Title, Link: it.Link, PublishedAt: it.PublishedAt, Read: it.Read})
	}
	return out, nil
}

// MarkRead implements Reader.
func (uc *reader) MarkRead(ctx context.Context, itemID string, read bool) error {
	return uc.d.Repo.SetRead(ctx, itemID, read)
}

// MarkAllRead implements Reader.
func (uc *reader) MarkAllRead(ctx context.Context, feedID string) error {
	if feedID != "" {
		if _, err := uc.d.Repo.GetFeed(ctx, feedID); err != nil {
			return err
		}
	}
	return uc.d.Repo.MarkAllRead(ctx, feedID)
}

// RefreshAll implements Reader; it is run by the background job. A broken
// feed is recorded on the feed and does not stop the others.
func (uc *reader) RefreshAll(ctx context.Context) error {
	feeds, err := uc.d.Repo.ListFeeds(ctx)
	if err != nil {
		return fmt.Errorf("list feeds: %w", err)
	}
	var failed []string
	for _, f := range feeds {
		fetched, err := uc.d.Fetcher.Fetch(ctx, f.URL)
		if err != nil {
			f.RecordFailure(err, uc.d.Clock.Now())
			failed = append(failed, f.URL)
			if uerr := uc.d.Repo.UpdateFeed(ctx, f); uerr != nil {
				return fmt.Errorf("record feed failure: %w", uerr)
			}
			continue
		}
		if err := uc.store(ctx, &f, fetched); err != nil {
			return err
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("refresh feeds %s: %w", strings.Join(failed, ", "), kernel.ErrUnavailable)
	}
	return nil
}

func (uc *reader) store(ctx context.Context, f *domain.Feed, fetched FetchedFeed) error {
	now := uc.d.Clock.Now()
	items := make([]domain.Item, 0, len(fetched.Items))
	for _, it := range fetched.Items {
		published := it.Published
		if published.IsZero() {
			published = now
		}
		items = append(items, domain.NewItem(uc.d.IDs.NewID(), f.ID, it.GUID, it.Title, it.Link, published))
	}
	if _, err := uc.d.Repo.AddItems(ctx, items); err != nil {
		return fmt.Errorf("add items: %w", err)
	}
	if err := uc.d.Repo.PruneItems(ctx, f.ID, domain.MaxItemsPerFeed); err != nil {
		return fmt.Errorf("prune items: %w", err)
	}
	f.RecordSuccess(fetched.Title, fetched.SiteURL, now)
	if err := uc.d.Repo.UpdateFeed(ctx, *f); err != nil {
		return fmt.Errorf("update feed: %w", err)
	}
	return nil
}

func (uc *reader) view(ctx context.Context, f domain.Feed) (FeedView, error) {
	counts, err := uc.d.Repo.UnreadCounts(ctx)
	if err != nil {
		return FeedView{}, fmt.Errorf("count unread: %w", err)
	}
	return toFeedView(f, counts[f.ID]), nil
}

func toFeedView(f domain.Feed, unread int) FeedView {
	return FeedView{ID: f.ID, URL: f.URL, Title: f.Title, SiteURL: f.SiteURL, LastError: f.LastError, LastFetchedAt: f.LastFetchedAt,
		Unread: unread, Category: f.Category}
}

// publicReason shortens fetch errors to something safe and useful to show.
func publicReason(err error) string {
	var pe *kernel.PublicError
	if errors.As(err, &pe) {
		return pe.Message
	}
	msg := err.Error()
	if len(msg) > 160 {
		msg = msg[:160] + "…"
	}
	return msg
}

// ErrNotOPML is returned for files that are not OPML.
var ErrNotOPML = kernel.NewPublicError(kernel.ErrInvalid, "The file is not an OPML subscription list.")

// ImportOPML implements Reader.
func (uc *reader) ImportOPML(ctx context.Context, data []byte) (ImportReport, error) {
	if len(data) > domain.MaxImportBytes {
		return ImportReport{}, kernel.NewValidationError("file", "The OPML file is larger than 2 MB.")
	}
	entries, err := uc.d.OPML.Decode(data)
	if err != nil {
		return ImportReport{}, ErrNotOPML
	}
	existing, err := uc.d.Repo.ListFeeds(ctx)
	if err != nil {
		return ImportReport{}, fmt.Errorf("list feeds: %w", err)
	}
	plan := domain.PlanImport(existing, entries)
	report := ImportReport{Duplicates: plan.Duplicates, Invalid: plan.Invalid}
	now := uc.d.Clock.Now()
	for _, e := range plan.Add {
		f, err := domain.NewFeed(uc.d.IDs.NewID(), e.URL, now)
		if err != nil {
			report.Invalid = append(report.Invalid, domain.InvalidEntry{Title: e.Title, URL: e.URL, Reason: err.Error()})
			continue
		}
		if e.Title != "" {
			f.Title = e.Title
		}
		f.SiteURL, f.Category = e.SiteURL, e.Category
		switch err := uc.d.Repo.CreateFeed(ctx, f); {
		case errors.Is(err, domain.ErrAlreadySubscribed):
			report.Duplicates++
		case err != nil:
			return report, fmt.Errorf("create feed: %w", err)
		default:
			report.Added++
		}
	}
	return report, nil
}

// ExportOPML implements Reader.
func (uc *reader) ExportOPML(ctx context.Context) ([]byte, error) {
	feeds, err := uc.d.Repo.ListFeeds(ctx)
	if err != nil {
		return nil, fmt.Errorf("list feeds: %w", err)
	}
	return uc.d.OPML.Encode(feeds, uc.d.Clock.Now())
}
