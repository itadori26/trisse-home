package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"slices"
	"strings"
	"time"
)

const migrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
	module     TEXT NOT NULL,
	version    TEXT NOT NULL,
	applied_at TEXT NOT NULL,
	PRIMARY KEY (module, version)
)`

// Migrate applies all *.sql files in fsys that have not yet been applied for
// module. Files are applied in lexical order (use 0001_name.sql prefixes),
// each inside its own transaction.
func Migrate(ctx context.Context, db *sql.DB, module string, fsys fs.FS) error {
	if fsys == nil {
		return nil
	}
	if _, err := db.ExecContext(ctx, migrationsTable); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}
	files, err := fs.Glob(fsys, "*.sql")
	if err != nil {
		return fmt.Errorf("list migrations of %s: %w", module, err)
	}
	slices.Sort(files)
	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		if err := applyOne(ctx, db, fsys, module, name, version); err != nil {
			return err
		}
	}
	return nil
}

func applyOne(ctx context.Context, db *sql.DB, fsys fs.FS, module, name, version string) (err error) {
	var exists int
	err = db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations WHERE module = ? AND version = ?`, module, version).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check migration %s/%s: %w", module, version, err)
	}
	if exists > 0 {
		return nil
	}
	body, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("read migration %s/%s: %w", module, name, err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s/%s: %w", module, version, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if _, err = tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("apply migration %s/%s: %w", module, version, err)
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (module, version, applied_at) VALUES (?, ?, ?)`,
		module, version, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("record migration %s/%s: %w", module, version, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s/%s: %w", module, version, err)
	}
	return nil
}
