package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"trisse-home/internal/modules/drop/domain"
)

type memRepo struct {
	items map[string]domain.Item
	files map[string][]byte
}

func (m *memRepo) List(context.Context) ([]domain.Item, error) {
	var out []domain.Item
	for _, i := range m.items {
		out = append(out, i)
	}
	return out, nil
}
func (m *memRepo) Get(_ context.Context, id string) (domain.Item, error) {
	i, ok := m.items[id]
	if !ok {
		return i, domain.ErrNotFound
	}
	return i, nil
}
func (m *memRepo) Add(_ context.Context, i domain.Item, data []byte) error {
	m.items[i.ID] = i
	if data != nil {
		m.files[i.ID] = data
	}
	return nil
}
func (m *memRepo) Delete(_ context.Context, id string) error {
	delete(m.items, id)
	delete(m.files, id)
	return nil
}
func (m *memRepo) File(_ context.Context, id string) ([]byte, error) {
	d, ok := m.files[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return d, nil
}
func (m *memRepo) DeleteExpired(_ context.Context, now time.Time) (int, error) {
	n := 0
	for id, i := range m.items {
		if i.Expired(now) {
			delete(m.items, id)
			delete(m.files, id)
			n++
		}
	}
	return n, nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "d" + strconv.Itoa(i.n) }

func TestDrop(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{items: map[string]domain.Item{}, files: map[string][]byte{}}
	c := &clk{t: time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)}
	uc := New(repo, c, &ids{})
	short, _ := uc.AddText(ctx, "kurz", time.Hour)
	c.t = c.t.Add(time.Minute)
	f, err := uc.AddFile(ctx, "Foto.jpg", "image/jpeg", []byte("jpg"), 0)
	if err != nil || f.Size != 3 || f.FileName != "Foto.jpg" {
		t.Fatalf("%+v %v", f, err)
	}
	if list, _ := uc.List(ctx); len(list) != 2 || list[0].ID != f.ID {
		t.Fatalf("newest first: %+v", list)
	}
	if _, data, err := uc.File(ctx, f.ID); err != nil || string(data) != "jpg" {
		t.Fatal(err)
	}
	if _, _, err := uc.File(ctx, short.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("text has no file")
	}
	// After an hour the short item is gone everywhere, even before the purge.
	c.t = c.t.Add(time.Hour)
	if list, _ := uc.List(ctx); len(list) != 1 {
		t.Fatalf("expired listed: %+v", list)
	}
	_ = uc.PurgeExpired(ctx)
	if len(repo.items) != 1 {
		t.Fatalf("purge: %d", len(repo.items))
	}
	if err := uc.Delete(ctx, f.ID); err != nil || len(repo.files) != 0 {
		t.Fatal(err)
	}
}
