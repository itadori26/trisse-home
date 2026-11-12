package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestoreLifecycle(t *testing.T) {
	dir := t.TempDir()
	db := filepath.Join(dir, "trisse-home.db")
	_ = os.WriteFile(db, []byte("old"), 0o600)
	_ = os.WriteFile(db+"-wal", []byte("old-wal"), 0o600)

	if applied, err := ApplyPendingRestore(db); applied || err != nil {
		t.Fatalf("nothing pending: %v %v", applied, err)
	}
	r := NewRestorer(db, 10)
	if _, err := r.Stage(context.Background(), strings.NewReader("way too large")); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("size limit: %v", err)
	}
	staged, err := r.Stage(context.Background(), strings.NewReader("new"))
	if err != nil || filepath.Dir(staged) != dir {
		t.Fatalf("stage: %q %v", staged, err)
	}
	if err := r.Schedule(filepath.Join(dir, "other")); err == nil {
		t.Fatal("foreign files must not be scheduled")
	}
	if err := r.Schedule(staged); err != nil {
		t.Fatal(err)
	}
	applied, err := ApplyPendingRestore(db)
	if err != nil || !applied {
		t.Fatalf("apply: %v %v", applied, err)
	}
	read := func(p string) string { b, _ := os.ReadFile(p); return string(b) }
	if read(db) != "new" || read(db+".before-restore") != "old" || read(db+".before-restore-wal") != "old-wal" {
		t.Fatal("swap result wrong")
	}
	if _, err := os.Stat(db + "-wal"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("old WAL must not stay next to the restored database")
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), stagePrefix) {
			t.Fatalf("left staging file %s", e.Name())
		}
	}
}

func TestDiscard(t *testing.T) {
	dir := t.TempDir()
	r := NewRestorer(filepath.Join(dir, "x.db"), 100)
	staged, _ := r.Stage(context.Background(), strings.NewReader("data"))
	keep := filepath.Join(dir, "keep")
	_ = os.WriteFile(keep, nil, 0o600)
	r.Discard(staged)
	r.Discard(keep)
	if _, err := os.Stat(staged); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("staged file not removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatal("foreign file removed")
	}
}
