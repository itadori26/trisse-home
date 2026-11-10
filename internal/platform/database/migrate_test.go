package database

import (
	"context"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestMigrateAppliesOnce(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	fsys := fstest.MapFS{
		"0001_init.sql": {Data: []byte(`CREATE TABLE t (id INTEGER PRIMARY KEY);`)},
		"0002_seed.sql": {Data: []byte(`INSERT INTO t (id) VALUES (1);`)},
	}
	for range 2 {
		if err := Migrate(ctx, db, "test", fsys); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("rows = %d, err = %v; want exactly one seeded row", n, err)
	}
}

func TestMigrateRollsBackFailure(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	fsys := fstest.MapFS{"0001_bad.sql": {Data: []byte(`CREATE TABLE a (id INT); SYNTAX ERROR;`)}}
	if err := Migrate(ctx, db, "bad", fsys); err == nil {
		t.Fatal("expected error")
	}
	var n int
	_ = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE name = 'a'`).Scan(&n)
	if n != 0 {
		t.Fatal("failed migration was not rolled back")
	}
}
