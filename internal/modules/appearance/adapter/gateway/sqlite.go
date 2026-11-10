// Package gateway stores the appearance in SQLite.
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

	"trisse-home/internal/modules/appearance/domain"
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

// Load returns the stored appearance or the default.
func (r *Repository) Load(ctx context.Context) (domain.Appearance, error) {
	var theme, accent, bg, css string
	err := r.db.QueryRowContext(ctx,
		`SELECT theme, accent_color, background_image, custom_css FROM appearance_settings WHERE id = 1`).
		Scan(&theme, &accent, &bg, &css)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Default(), nil
	}
	if err != nil {
		return domain.Appearance{}, fmt.Errorf("select appearance: %w", err)
	}
	a, err := domain.NewAppearance(domain.Theme(theme), accent, bg, css)
	if err != nil {
		// Stored data that no longer satisfies the rules falls back to the
		// default instead of locking the user out of the dashboard.
		return domain.Default(), nil //nolint:nilerr // intentional fallback
	}
	return a, nil
}

// Save stores the appearance.
func (r *Repository) Save(ctx context.Context, a domain.Appearance) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO appearance_settings (id, theme, accent_color, background_image, custom_css)
		VALUES (1, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET theme = excluded.theme, accent_color = excluded.accent_color,
			background_image = excluded.background_image, custom_css = excluded.custom_css`,
		string(a.Theme()), a.AccentColor(), a.BackgroundImage(), a.CustomCSS())
	if err != nil {
		return fmt.Errorf("upsert appearance: %w", err)
	}
	return nil
}
