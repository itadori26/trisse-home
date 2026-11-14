// Package gateway stores summaries in SQLite, asks Ollama and loads articles.
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

	"trisse-home/internal/modules/summary/domain"
	"trisse-home/internal/modules/summary/usecase"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are exported: the settings hold no secret; the cache is
// rebuilt on demand.
var TransferTables = []sqlutil.Table{{Name: "summary_settings"}}

// MaxCached bounds the cache; the oldest summaries go first.
const MaxCached = 500

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Load implements usecase.Repository.
func (r *Repository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	err := r.db.QueryRowContext(ctx, `SELECT url, model, max_input_chars FROM summary_settings WHERE id = 1`).Scan(&s.URL, &s.Model, &s.MaxInputChars)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	return s, nil
}

// Save implements usecase.Repository.
func (r *Repository) Save(ctx context.Context, s domain.Settings) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO summary_settings (id, url, model, max_input_chars) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, model = excluded.model, max_input_chars = excluded.max_input_chars`,
		s.URL, s.Model, s.MaxInputChars)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

// Clear implements usecase.Repository: settings and summaries go.
func (r *Repository) Clear(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM summary_settings`); err != nil {
		return fmt.Errorf("delete settings: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM summary_cache`); err != nil {
		return fmt.Errorf("delete cache: %w", err)
	}
	return nil
}

// Cached implements usecase.Repository.
func (r *Repository) Cached(ctx context.Context, key string) (usecase.Cached, bool, error) {
	var c usecase.Cached
	var created string
	err := r.db.QueryRowContext(ctx, `SELECT text, model, created_at FROM summary_cache WHERE key = ?`, key).Scan(&c.Text, &c.Model, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.Cached{}, false, nil
	}
	if err != nil {
		return usecase.Cached{}, false, fmt.Errorf("select summary: %w", err)
	}
	c.CreatedAt, err = sqlutil.ParseTime(created)
	return c, err == nil, err
}

// Store implements usecase.Repository and keeps at most MaxCached rows.
func (r *Repository) Store(ctx context.Context, key string, c usecase.Cached) error {
	if _, err := r.db.ExecContext(ctx, `INSERT INTO summary_cache (key, text, model, created_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (key) DO UPDATE SET text = excluded.text, model = excluded.model, created_at = excluded.created_at`,
		key, c.Text, c.Model, sqlutil.FormatTime(c.CreatedAt)); err != nil {
		return fmt.Errorf("insert summary: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM summary_cache WHERE key NOT IN
		(SELECT key FROM summary_cache ORDER BY created_at DESC LIMIT ?)`, MaxCached); err != nil {
		return fmt.Errorf("prune summaries: %w", err)
	}
	return nil
}
