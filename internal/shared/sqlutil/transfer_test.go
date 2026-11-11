package sqlutil_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"testing/fstest"

	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/sqlutil"
)

func TestDumpAndRestore(t *testing.T) {
	ctx := context.Background()
	migrations := fstest.MapFS{"0001_init.sql": {Data: []byte(`
		CREATE TABLE parents (id TEXT PRIMARY KEY, name TEXT NOT NULL, secret TEXT NOT NULL DEFAULT '');
		CREATE TABLE children (id TEXT PRIMARY KEY, parent_id TEXT NOT NULL REFERENCES parents (id) ON DELETE CASCADE, n INTEGER NOT NULL, f REAL);`)}}
	db := dbtest.Open(t, "transfer", migrations)
	_, _ = db.ExecContext(ctx, `INSERT INTO parents VALUES ('p1', 'Home', 'pw'); INSERT INTO children VALUES ('c1', 'p1', 3, 1.5)`)
	tables := []sqlutil.Table{{Name: "parents", Exclude: []string{"secret"}}, {Name: "children"}}

	dump, err := sqlutil.DumpTables(ctx, db, tables)
	if err != nil || len(dump["parents"]) != 1 || dump["parents"][0]["secret"] != nil {
		t.Fatalf("dump %v %v", dump, err)
	}
	raw, _ := json.Marshal(dump)
	var back map[string]sqlutil.Rows
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	_ = dec.Decode(&back)
	back["children"][0]["unknown_column"] = "ignored"
	if err := sqlutil.RestoreTables(ctx, db, tables, back); err != nil {
		t.Fatal(err)
	}
	var n int
	var f float64
	var secret string
	_ = db.QueryRowContext(ctx, `SELECT n, f FROM children`).Scan(&n, &f)
	_ = db.QueryRowContext(ctx, `SELECT secret FROM parents`).Scan(&secret)
	if n != 3 || f != 1.5 || secret != "" {
		t.Fatalf("restored n=%d f=%v secret=%q", n, f, secret)
	}

	broken := map[string]sqlutil.Rows{"children": {{"id": "c2", "parent_id": "missing", "n": json.Number("1")}}}
	if err := sqlutil.RestoreTables(ctx, db, tables, broken); !errors.Is(err, sqlutil.ErrInvalidExport) {
		t.Fatalf("broken references accepted: %v", err)
	}
	var count int
	_ = db.QueryRowContext(ctx, `SELECT count(*) FROM parents`).Scan(&count)
	if count != 1 {
		t.Fatal("failed import must roll back")
	}
	if err := sqlutil.RestoreTables(ctx, db, tables, map[string]sqlutil.Rows{"users": {}}); !errors.Is(err, sqlutil.ErrInvalidExport) {
		t.Fatal("foreign table accepted")
	}
}
