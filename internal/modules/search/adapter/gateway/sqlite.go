package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/modules/search/usecase"
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

// DefaultProvider implements usecase.Repository.
func (r *Repository) DefaultProvider(ctx context.Context) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT default_provider FROM search_settings WHERE id = 1`).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("select default provider: %w", err)
	}
	return id, nil
}

// SetDefaultProvider implements usecase.Repository.
func (r *Repository) SetDefaultProvider(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO search_settings (id, default_provider) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET default_provider = excluded.default_provider`, id)
	if err != nil {
		return fmt.Errorf("upsert default provider: %w", err)
	}
	return nil
}

// BarOptions implements usecase.Repository.
func (r *Repository) BarOptions(ctx context.Context) (domain.BarOptions, error) {
	var newTab bool
	var raw string
	err := r.db.QueryRowContext(ctx, `SELECT open_in_new_tab, capture_prefixes FROM search_settings WHERE id = 1`).Scan(&newTab, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultBarOptions(), nil
	}
	if err != nil {
		return domain.BarOptions{}, fmt.Errorf("select search field options: %w", err)
	}
	p := domain.DefaultCapturePrefixes
	if raw != "" {
		var stored struct{ Todo, Note, Bookmark string }
		if json.Unmarshal([]byte(raw), &stored) == nil {
			p = domain.CapturePrefixes{Todo: stored.Todo, Note: stored.Note, Bookmark: stored.Bookmark}
		}
	}
	o, err := domain.NewBarOptions(newTab, p)
	if err != nil {
		// A row that violates current rules falls back to the default prefixes
		// instead of breaking the search field.
		o, _ = domain.NewBarOptions(newTab, domain.DefaultCapturePrefixes)
	}
	return o, nil
}

// SetBarOptions implements usecase.Repository.
func (r *Repository) SetBarOptions(ctx context.Context, o domain.BarOptions) error {
	p := o.Prefixes()
	raw, err := json.Marshal(struct{ Todo, Note, Bookmark string }{p.Todo, p.Note, p.Bookmark})
	if err != nil {
		return fmt.Errorf("encode capture prefixes: %w", err)
	}
	// No row yet: default_provider '' means "the default provider".
	_, err = r.db.ExecContext(ctx, `INSERT INTO search_settings (id, default_provider, open_in_new_tab, capture_prefixes)
		VALUES (1, '', ?, ?)
		ON CONFLICT (id) DO UPDATE SET open_in_new_tab = excluded.open_in_new_tab, capture_prefixes = excluded.capture_prefixes`,
		o.OpenInNewTab(), string(raw))
	if err != nil {
		return fmt.Errorf("upsert search field options: %w", err)
	}
	return nil
}

// Shortcuts implements usecase.Repository.
func (r *Repository) Shortcuts(ctx context.Context) ([]domain.Shortcut, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT keyword, name, url_template FROM search_shortcuts ORDER BY keyword`)
	if err != nil {
		return nil, fmt.Errorf("query shortcuts: %w", err)
	}
	defer rows.Close()
	var out []domain.Shortcut
	for rows.Next() {
		var k, n, t string
		if err := rows.Scan(&k, &n, &t); err != nil {
			return nil, fmt.Errorf("scan shortcut: %w", err)
		}
		s, err := domain.NewShortcut(k, n, t)
		if err != nil {
			continue // skip rows that violate current rules instead of failing every search
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// AddShortcut implements usecase.Repository.
func (r *Repository) AddShortcut(ctx context.Context, s domain.Shortcut) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO search_shortcuts (keyword, name, url_template) VALUES (?, ?, ?)`,
		s.Keyword(), s.Name(), s.Template())
	if sqlutil.IsUniqueViolation(err) {
		return usecase.ErrShortcutExists
	}
	if err != nil {
		return fmt.Errorf("insert shortcut: %w", err)
	}
	return nil
}

// DeleteShortcut implements usecase.Repository.
func (r *Repository) DeleteShortcut(ctx context.Context, keyword string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM search_shortcuts WHERE keyword = ?`, keyword)
	if err != nil {
		return fmt.Errorf("delete shortcut: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return usecase.ErrShortcutNotFound
	}
	return nil
}
