// Package gateway stores countdowns in SQLite.
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
	"time"

	"trisse-home/internal/modules/countdown/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are the tables included in data exports (no secrets).
var TransferTables = []sqlutil.Table{{Name: "countdown_items"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, title, emoji, day, minute, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Countdown, error) {
	var c domain.Countdown
	var day, created string
	if err := sc.Scan(&c.ID, &c.Title, &c.Emoji, &day, &c.Minute, &created); err != nil {
		return c, err
	}
	var err error
	if c.Day, err = time.Parse(time.DateOnly, day); err != nil {
		return c, err
	}
	c.CreatedAt, err = sqlutil.ParseTime(created)
	return c, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Countdown, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM countdown_items ORDER BY day, minute`)
	if err != nil {
		return nil, fmt.Errorf("query countdowns: %w", err)
	}
	defer rows.Close()
	var out []domain.Countdown
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan countdown: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Countdown, error) {
	c, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM countdown_items WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Countdown{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Countdown{}, fmt.Errorf("select countdown: %w", err)
	}
	return c, nil
}

// Save implements usecase.Repository (upsert).
func (r *Repository) Save(ctx context.Context, c domain.Countdown) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO countdown_items (`+columns+`) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET title = excluded.title, emoji = excluded.emoji, day = excluded.day, minute = excluded.minute`,
		c.ID, c.Title, c.Emoji, c.Date(), c.Minute, sqlutil.FormatTime(c.CreatedAt))
	if err != nil {
		return fmt.Errorf("upsert countdown: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM countdown_items WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete countdown: %w", err)
	}
	return nil
}
