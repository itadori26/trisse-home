package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/backup/domain"
)

type memStore struct{ files map[string]domain.File }

func (m *memStore) Path(name string) string { return "/backups/" + name }
func (m *memStore) List(context.Context) ([]domain.File, error) {
	var out []domain.File
	for _, f := range m.files {
		out = append(out, f)
	}
	slices.SortFunc(out, func(a, b domain.File) int { return a.CreatedAt.Compare(b.CreatedAt) })
	return out, nil
}
func (m *memStore) Open(name string) (io.ReadCloser, domain.File, error) {
	f, ok := m.files[name]
	if !ok {
		return nil, domain.File{}, domain.ErrBackupNotFound
	}
	return io.NopCloser(bytes.NewReader([]byte("data"))), f, nil
}
func (m *memStore) Remove(name string) error { delete(m.files, name); return nil }

type fakeSnap struct {
	store *memStore
	clock *clk
	err   error
}

func (s *fakeSnap) SnapshotTo(_ context.Context, path string) error {
	if s.err != nil {
		return s.err
	}
	name := strings.TrimPrefix(path, "/backups/")
	s.store.files[name] = domain.File{Name: name, Size: 4, CreatedAt: s.clock.t}
	return nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

func TestBackups(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC)}
	store := &memStore{files: map[string]domain.File{}}
	policy, _ := domain.NewPolicy(24*time.Hour, 2)
	uc := NewBackups(policy, &fakeSnap{store: store, clock: c}, store, c)

	for range 3 { // first run creates, the next two are not due
		if err := uc.RunScheduled(ctx); err != nil {
			t.Fatal(err)
		}
		c.t = c.t.Add(time.Hour)
	}
	if len(store.files) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(store.files))
	}
	for range 3 {
		c.t = c.t.Add(24 * time.Hour)
		if _, err := uc.BackupNow(ctx); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := uc.List(ctx)
	if len(list) != 2 || !list[0].CreatedAt.After(list[1].CreatedAt) {
		t.Fatalf("retention/order wrong: %+v", list)
	}
	if _, _, err := uc.Open(ctx, "../../etc/passwd"); !errors.Is(err, domain.ErrBackupNotFound) {
		t.Fatalf("path traversal not rejected: %v", err)
	}
	rc, f, err := uc.Open(ctx, list[0].Name)
	if err != nil || f.Size != 4 {
		t.Fatalf("open: %v", err)
	}
	_ = rc.Close()
}
