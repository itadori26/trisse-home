package gateway

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/platform/database"
	"trisse-home/internal/platform/database/dbtest"
)

func TestSnapshotAndStore(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "test", fstest.MapFS{"0001.sql": {Data: []byte(`CREATE TABLE t (v TEXT); INSERT INTO t VALUES ('hello');`)}})
	store, err := NewDirStore(filepath.Join(t.TempDir(), "backups"))
	if err != nil {
		t.Fatal(err)
	}
	name := domain.FileName(time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC))
	if err := NewSnapshotter(db).SnapshotTo(ctx, store.Path(name)); err != nil {
		t.Fatal(err)
	}
	files, err := store.List(ctx)
	if err != nil || len(files) != 1 || files[0].Name != name || files[0].Size == 0 || files[0].CreatedAt.Hour() != 3 {
		t.Fatalf("list %+v %v", files, err)
	}
	// The backup is a working database with the original content.
	copyDB, err := database.Open(ctx, store.Path(name))
	if err != nil {
		t.Fatal(err)
	}
	var v string
	if err := copyDB.QueryRowContext(ctx, `SELECT v FROM t`).Scan(&v); err != nil || v != "hello" {
		t.Fatalf("backup content %q %v", v, err)
	}
	_ = copyDB.Close()
	rc, _, err := store.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	head := make([]byte, 15)
	_, _ = io.ReadFull(rc, head)
	_ = rc.Close()
	if string(head) != "SQLite format 3" {
		t.Fatalf("not an SQLite file: %q", head)
	}
	if _, _, err := store.Open("../x.db"); err == nil {
		t.Fatal("invalid name accepted")
	}
	if err := store.Remove(name); err != nil {
		t.Fatal(err)
	}
}
