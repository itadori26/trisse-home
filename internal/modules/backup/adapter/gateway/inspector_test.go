package gateway

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestInspector(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "backup", Migrations())
	_, _ = db.Exec(`INSERT INTO schema_migrations (module, version, applied_at) VALUES ('auth', '0001', '2026-01-01T00:00:00Z')`)
	insp := NewInspector(db)
	dir := t.TempDir()

	snapshot := filepath.Join(dir, "snap.db")
	if err := NewSnapshotter(db).SnapshotTo(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	info, err := insp.Inspect(ctx, snapshot)
	if err != nil || info.Integrity != "ok" || len(info.Migrations) != 2 {
		t.Fatalf("snapshot: %+v %v", info, err)
	}
	current, _ := insp.Current(ctx)
	if err := domain.ValidateRestore(info, current); err != nil {
		t.Fatalf("own snapshot must be restorable: %v", err)
	}

	garbage := filepath.Join(dir, "garbage.db")
	_ = os.WriteFile(garbage, []byte("this is not sqlite at all, just some text that is long enough"), 0o600)
	if _, err := insp.Inspect(ctx, garbage); err == nil {
		t.Fatal("garbage accepted")
	}

	// An empty SQLite database has no schema_migrations table.
	empty := filepath.Join(dir, "empty.db")
	_, _ = db.Exec(`VACUUM INTO ?`, filepath.Join(dir, "tmp.db"))
	other := dbtest.Open(t, "x", nil)
	_, _ = other.Exec(`DROP TABLE schema_migrations`)
	_, _ = other.Exec(`VACUUM INTO ?`, empty)
	info, err = insp.Inspect(ctx, empty)
	if err != nil || len(info.Migrations) != 0 || domain.ValidateRestore(info, current) == nil {
		t.Fatalf("foreign db: %+v %v", info, err)
	}
	// The live database must still work and have nothing attached.
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_database_list WHERE name = 'restore_check'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("still attached: %d %v", n, err)
	}
}
