// Package dbtest opens throw-away SQLite databases for gateway integration
// tests.
//
// Ring: Frameworks & Drivers (test support). Only _test.go files may import it.
package dbtest

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"testing"

	"trisse-home/internal/platform/database"
)

// Open creates a temporary database with the given module migrations applied.
func Open(t testing.TB, module string, migrations fs.FS) *sql.DB {
	t.Helper()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := database.Migrate(ctx, db, module, migrations); err != nil {
		t.Fatalf("migrate %s: %v", module, err)
	}
	return db
}
