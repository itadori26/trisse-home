// Package gateway stores focus sessions in SQLite.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"time"

	"trisse-home/internal/modules/focus/domain"
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
var TransferTables = []sqlutil.Table{{Name: "focus_sessions"}}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// Save implements usecase.Repository.
func (r *Repository) Save(ctx context.Context, s domain.Session) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO focus_sessions (id, minutes, ended_at) VALUES (?, ?, ?)`,
		s.ID, s.Minutes, sqlutil.FormatTime(s.EndedAt))
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// Since implements usecase.Repository. Stored times are UTC with a fixed
// width, so they compare correctly as text.
func (r *Repository) Since(ctx context.Context, t time.Time) ([]domain.Session, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, minutes, ended_at FROM focus_sessions WHERE ended_at >= ? ORDER BY ended_at`,
		sqlutil.FormatTime(t))
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()
	var out []domain.Session
	for rows.Next() {
		var s domain.Session
		var ended string
		if err := rows.Scan(&s.ID, &s.Minutes, &ended); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		if s.EndedAt, err = sqlutil.ParseTime(ended); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeleteBefore implements usecase.Repository.
func (r *Repository) DeleteBefore(ctx context.Context, t time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM focus_sessions WHERE ended_at < ?`, sqlutil.FormatTime(t)); err != nil {
		return fmt.Errorf("prune sessions: %w", err)
	}
	return nil
}
