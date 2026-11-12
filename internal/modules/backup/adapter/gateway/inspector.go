package gateway

import (
	"context"
	"database/sql"
	"fmt"

	"trisse-home/internal/modules/backup/domain"
)

// Inspector implements usecase.BackupInspector. It attaches the candidate
// file to a dedicated connection of the live database, so no second driver
// connection is opened and nothing is written to the file.
type Inspector struct{ db *sql.DB }

// NewInspector returns an inspector using db.
func NewInspector(db *sql.DB) *Inspector { return &Inspector{db: db} }

// Inspect implements usecase.BackupInspector.
func (i *Inspector) Inspect(ctx context.Context, path string) (domain.BackupInfo, error) {
	conn, err := i.db.Conn(ctx)
	if err != nil {
		return domain.BackupInfo{}, fmt.Errorf("get connection: %w", err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `ATTACH DATABASE ? AS restore_check`, path); err != nil {
		return domain.BackupInfo{}, fmt.Errorf("attach backup: %w", err)
	}
	defer func() { _, _ = conn.ExecContext(context.WithoutCancel(ctx), `DETACH DATABASE restore_check`) }()
	var info domain.BackupInfo
	// The first row is "ok" for a sound database.
	if err := conn.QueryRowContext(ctx, `PRAGMA restore_check.integrity_check`).Scan(&info.Integrity); err != nil {
		return domain.BackupInfo{}, fmt.Errorf("integrity check: %w", err)
	}
	var tables int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM restore_check.sqlite_master WHERE type = 'table' AND name = 'schema_migrations'`).Scan(&tables); err != nil {
		return domain.BackupInfo{}, fmt.Errorf("read schema: %w", err)
	}
	if tables == 0 {
		return info, nil
	}
	info.Migrations, err = migrations(ctx, conn, `SELECT module, version FROM restore_check.schema_migrations`)
	return info, err
}

// Current implements usecase.BackupInspector.
func (i *Inspector) Current(ctx context.Context) ([]domain.Migration, error) {
	conn, err := i.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	defer conn.Close()
	return migrations(ctx, conn, `SELECT module, version FROM main.schema_migrations`)
}

func migrations(ctx context.Context, conn *sql.Conn, query string) ([]domain.Migration, error) {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	defer rows.Close()
	var out []domain.Migration
	for rows.Next() {
		var m domain.Migration
		if err := rows.Scan(&m.Module, &m.Version); err != nil {
			return nil, fmt.Errorf("scan migration: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
