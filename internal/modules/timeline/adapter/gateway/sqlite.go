// Package gateway stores timeline events in SQLite.
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

	"trisse-home/internal/modules/timeline/domain"
	"trisse-home/internal/modules/timeline/usecase"
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

// Add implements usecase.Repository.
func (r *Repository) Add(ctx context.Context, e domain.Event) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO timeline_events (id, at, source, severity, subject, message) VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, sqlutil.FormatTime(e.At), e.Source, string(e.Severity), e.Subject, e.Message)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

// List implements usecase.Repository (newest first).
func (r *Repository) List(ctx context.Context, f usecase.Filter) ([]domain.Event, error) {
	before := "9999"
	if !f.Before.IsZero() {
		before = sqlutil.FormatTime(f.Before)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, at, source, severity, subject, message FROM timeline_events
		WHERE at < ? AND (? = '' OR source = ?) AND (? = '' OR severity = ?) ORDER BY at DESC, id DESC LIMIT ?`,
		before, f.Source, f.Source, string(f.Severity), string(f.Severity), f.Limit)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()
	out := []domain.Event{}
	for rows.Next() {
		var e domain.Event
		var at, sev string
		if err := rows.Scan(&e.ID, &at, &e.Source, &sev, &e.Subject, &e.Message); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Severity = domain.Severity(sev)
		if e.At, err = sqlutil.ParseTime(at); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// DeleteBefore implements usecase.Repository.
func (r *Repository) DeleteBefore(ctx context.Context, t time.Time) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM timeline_events WHERE at < ?`, sqlutil.FormatTime(t))
	return err
}

// RetentionDays implements usecase.Repository.
func (r *Repository) RetentionDays(ctx context.Context) (int, error) {
	var d int
	err := r.db.QueryRowContext(ctx, `SELECT retention_days FROM timeline_settings WHERE id = 1`).Scan(&d)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return d, err
}

// SetRetentionDays implements usecase.Repository.
func (r *Repository) SetRetentionDays(ctx context.Context, days int) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO timeline_settings (id, retention_days) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET retention_days = excluded.retention_days`, days)
	return err
}
