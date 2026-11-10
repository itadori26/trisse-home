// Package gateway implements the feed reader output ports: the SQLite
// repository and the HTTP feed fetcher (gofeed).
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/rss/domain"
	"trisse-home/internal/modules/rss/usecase"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const feedColumns = `id, url, title, site_url, last_fetched_at, last_error, created_at`

func scanFeed(s interface{ Scan(...any) error }) (domain.Feed, error) {
	var f domain.Feed
	var fetched *string
	var created string
	if err := s.Scan(&f.ID, &f.URL, &f.Title, &f.SiteURL, &fetched, &f.LastError, &created); err != nil {
		return f, err
	}
	var err error
	if f.LastFetchedAt, err = sqlutil.ParseNullableTime(fetched); err != nil {
		return f, err
	}
	f.CreatedAt, err = sqlutil.ParseTime(created)
	return f, err
}

// ListFeeds returns all feeds ordered by title.
func (r *Repository) ListFeeds(ctx context.Context) ([]domain.Feed, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+feedColumns+` FROM rss_feeds ORDER BY title COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("query feeds: %w", err)
	}
	defer rows.Close()
	var out []domain.Feed
	for rows.Next() {
		f, err := scanFeed(rows)
		if err != nil {
			return nil, fmt.Errorf("scan feed: %w", err)
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// GetFeed returns one feed.
func (r *Repository) GetFeed(ctx context.Context, id string) (domain.Feed, error) {
	f, err := scanFeed(r.db.QueryRowContext(ctx, `SELECT `+feedColumns+` FROM rss_feeds WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Feed{}, domain.ErrFeedNotFound
	}
	if err != nil {
		return domain.Feed{}, fmt.Errorf("select feed: %w", err)
	}
	return f, nil
}

// CreateFeed inserts a feed.
func (r *Repository) CreateFeed(ctx context.Context, f domain.Feed) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO rss_feeds (`+feedColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		f.ID, f.URL, f.Title, f.SiteURL, sqlutil.FormatNullableTime(f.LastFetchedAt), f.LastError, sqlutil.FormatTime(f.CreatedAt))
	if sqlutil.IsUniqueViolation(err) {
		return domain.ErrAlreadySubscribed
	}
	if err != nil {
		return fmt.Errorf("insert feed: %w", err)
	}
	return nil
}

// UpdateFeed stores the fetch metadata of a feed.
func (r *Repository) UpdateFeed(ctx context.Context, f domain.Feed) error {
	_, err := r.db.ExecContext(ctx, `UPDATE rss_feeds SET title = ?, site_url = ?, last_fetched_at = ?, last_error = ? WHERE id = ?`,
		f.Title, f.SiteURL, sqlutil.FormatNullableTime(f.LastFetchedAt), f.LastError, f.ID)
	if err != nil {
		return fmt.Errorf("update feed: %w", err)
	}
	return nil
}

// DeleteFeed removes a feed; its items cascade.
func (r *Repository) DeleteFeed(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM rss_feeds WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete feed: %w", err)
	}
	return nil
}

// AddItems inserts new items and ignores known GUIDs (keeping their read state).
func (r *Repository) AddItems(ctx context.Context, items []domain.Item) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	inserted := 0
	for _, it := range items {
		res, err := tx.ExecContext(ctx, `INSERT INTO rss_items (id, feed_id, guid, title, link, published_at, read)
			VALUES (?, ?, ?, ?, ?, ?, 0) ON CONFLICT (feed_id, guid) DO UPDATE SET title = excluded.title, link = excluded.link`,
			it.ID, it.FeedID, it.GUID, it.Title, it.Link, sqlutil.FormatTime(it.PublishedAt))
		if err != nil {
			return 0, fmt.Errorf("insert item: %w", err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	}
	return inserted, tx.Commit()
}

// PruneItems keeps only the newest `keep` items of a feed.
func (r *Repository) PruneItems(ctx context.Context, feedID string, keep int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM rss_items WHERE feed_id = ? AND id NOT IN (
		SELECT id FROM rss_items WHERE feed_id = ? ORDER BY published_at DESC LIMIT ?)`, feedID, feedID, keep)
	if err != nil {
		return fmt.Errorf("prune items: %w", err)
	}
	return nil
}

// ListItems returns items, newest first.
func (r *Repository) ListItems(ctx context.Context, f usecase.ItemFilter) ([]domain.Item, error) {
	query := `SELECT id, feed_id, guid, title, link, published_at, read FROM rss_items WHERE 1 = 1`
	var args []any
	if f.FeedID != "" {
		query += ` AND feed_id = ?`
		args = append(args, f.FeedID)
	}
	if f.UnreadOnly {
		query += ` AND read = 0`
	}
	query += ` ORDER BY published_at DESC LIMIT ?`
	args = append(args, f.Limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		var it domain.Item
		var published string
		var read int
		if err := rows.Scan(&it.ID, &it.FeedID, &it.GUID, &it.Title, &it.Link, &published, &read); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		if it.PublishedAt, err = sqlutil.ParseTime(published); err != nil {
			return nil, err
		}
		it.Read = read == 1
		out = append(out, it)
	}
	return out, rows.Err()
}

// SetRead sets the read state of one item.
func (r *Repository) SetRead(ctx context.Context, itemID string, read bool) error {
	res, err := r.db.ExecContext(ctx, `UPDATE rss_items SET read = ? WHERE id = ?`, sqlutil.BoolToInt(read), itemID)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

// MarkAllRead marks all items, or those of one feed, as read.
func (r *Repository) MarkAllRead(ctx context.Context, feedID string) error {
	var err error
	if feedID == "" {
		_, err = r.db.ExecContext(ctx, `UPDATE rss_items SET read = 1 WHERE read = 0`)
	} else {
		_, err = r.db.ExecContext(ctx, `UPDATE rss_items SET read = 1 WHERE read = 0 AND feed_id = ?`, feedID)
	}
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

// UnreadCounts returns the number of unread items per feed.
func (r *Repository) UnreadCounts(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT feed_id, COUNT(*) FROM rss_items WHERE read = 0 GROUP BY feed_id`)
	if err != nil {
		return nil, fmt.Errorf("count unread: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, fmt.Errorf("scan count: %w", err)
		}
		out[id] = n
	}
	return out, rows.Err()
}
