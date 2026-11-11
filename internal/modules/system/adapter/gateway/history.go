package gateway

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"time"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// HistoryStore keeps metric points in SQLite.
type HistoryStore struct{ db *sql.DB }

// NewHistoryStore returns a store backed by db.
func NewHistoryStore(db *sql.DB) *HistoryStore { return &HistoryStore{db: db} }

// Add implements usecase.HistoryStore.
func (s *HistoryStore) Add(ctx context.Context, p domain.Point) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO system_history (at, cpu, memory, disk) VALUES (?, ?, ?, ?)`,
		sqlutil.FormatTime(p.At), p.CPU, p.Memory, p.Disk)
	if err != nil {
		return fmt.Errorf("insert history point: %w", err)
	}
	return nil
}

// Since implements usecase.HistoryStore; points are ordered by time.
func (s *HistoryStore) Since(ctx context.Context, t time.Time) ([]domain.Point, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT at, cpu, memory, disk FROM system_history WHERE at >= ? ORDER BY at`, sqlutil.FormatTime(t))
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()
	var out []domain.Point
	for rows.Next() {
		var at string
		var p domain.Point
		if err := rows.Scan(&at, &p.CPU, &p.Memory, &p.Disk); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		if p.At, err = sqlutil.ParseTime(at); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// PruneBefore implements usecase.HistoryStore.
func (s *HistoryStore) PruneBefore(ctx context.Context, t time.Time) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM system_history WHERE at < ?`, sqlutil.FormatTime(t)); err != nil {
		return fmt.Errorf("prune history: %w", err)
	}
	return nil
}
