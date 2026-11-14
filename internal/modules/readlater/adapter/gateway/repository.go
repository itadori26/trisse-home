// Package gateway stores articles in SQLite, fetches pages and images and
// extracts a sanitised reader copy.
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

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/modules/readlater/usecase"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are exported with the copies but without the images
// (binary data does not fit the JSON export; backups include them).
var TransferTables = []sqlutil.Table{{Name: "readlater_articles"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, url, title, site, excerpt, content, words, status, saved_at, bytes, error`

func scan(row interface{ Scan(...any) error }) (domain.Article, error) {
	var a domain.Article
	var status, saved string
	if err := row.Scan(&a.ID, &a.URL, &a.Title, &a.Site, &a.Excerpt, &a.Content, &a.Words, &status, &saved, &a.Bytes, &a.Error); err != nil {
		return domain.Article{}, err
	}
	a.Status = domain.Status(status)
	var err error
	a.SavedAt, err = sqlutil.ParseTime(saved)
	return a, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Article, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM readlater_articles`)
	if err != nil {
		return nil, fmt.Errorf("select articles: %w", err)
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		a, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Article, error) {
	a, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM readlater_articles WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Article{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Article{}, fmt.Errorf("select article: %w", err)
	}
	return a, nil
}

// Save implements usecase.Repository: the article and its images in one transaction.
func (r *Repository) Save(ctx context.Context, a domain.Article, images map[int]usecase.Image) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, `INSERT INTO readlater_articles (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.URL, a.Title, a.Site, a.Excerpt, a.Content, a.Words, string(a.Status), sqlutil.FormatTime(a.SavedAt), a.Bytes, a.Error); err != nil {
		return fmt.Errorf("insert article: %w", err)
	}
	for n, img := range images {
		if _, err = tx.ExecContext(ctx, `INSERT INTO readlater_images (article_id, n, content_type, data) VALUES (?, ?, ?, ?)`,
			a.ID, n, img.ContentType, img.Bytes); err != nil {
			return fmt.Errorf("insert image: %w", err)
		}
	}
	return tx.Commit()
}

// SetStatus implements usecase.Repository.
func (r *Repository) SetStatus(ctx context.Context, id string, s domain.Status) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE readlater_articles SET status = ? WHERE id = ?`, string(s), id); err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository (images go with the article).
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM readlater_images WHERE article_id = ?`, id); err != nil {
		return fmt.Errorf("delete images: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM readlater_articles WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete article: %w", err)
	}
	return nil
}

// Image implements usecase.Repository.
func (r *Repository) Image(ctx context.Context, id string, n int) (usecase.Image, error) {
	var img usecase.Image
	err := r.db.QueryRowContext(ctx, `SELECT content_type, data FROM readlater_images WHERE article_id = ? AND n = ?`, id, n).Scan(&img.ContentType, &img.Bytes)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Image{}, domain.ErrNotFound
	}
	if err != nil {
		return usecase.Image{}, fmt.Errorf("select image: %w", err)
	}
	return img, nil
}

// TotalBytes implements usecase.Repository.
func (r *Repository) TotalBytes(ctx context.Context) (int64, error) {
	var n sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT SUM(bytes) FROM readlater_articles`).Scan(&n); err != nil {
		return 0, fmt.Errorf("sum bytes: %w", err)
	}
	return n.Int64, nil
}
