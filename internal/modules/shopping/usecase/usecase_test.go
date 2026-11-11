package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type memStore struct {
	items map[string]domain.Item
	down  bool
	n     int
}

func (m *memStore) List(context.Context, domain.Settings) ([]domain.Item, error) {
	if m.down {
		return nil, errors.New("connection refused")
	}
	var out []domain.Item
	for _, it := range m.items {
		out = append(out, it)
	}
	return out, nil
}
func (m *memStore) Add(_ context.Context, _ domain.Settings, uid, summary string) (domain.Item, error) {
	m.n++
	it := domain.Item{ID: fmt.Sprintf("i%d.ics", m.n), Summary: summary, Created: time.Unix(int64(m.n), 0)}
	m.items[it.ID] = it
	return it, nil
}
func (m *memStore) SetDone(_ context.Context, _ domain.Settings, id string, done bool) error {
	it, ok := m.items[id]
	if !ok {
		return domain.ErrItemNotFound
	}
	it.Done = done
	m.items[id] = it
	return nil
}
func (m *memStore) Delete(_ context.Context, _ domain.Settings, id string) error {
	delete(m.items, id)
	return nil
}

type ids struct{}

func (ids) NewID() string { return "x" }

func TestShopping(t *testing.T) {
	ctx := context.Background()
	repo, store := &memRepo{}, &memStore{items: map[string]domain.Item{}}
	uc := NewShopping(repo, store, ids{})
	if _, err := uc.List(ctx); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := uc.SaveSettings(ctx, "https://c/cal/shopping/", "me", "pw"); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Add(ctx, "  "); err == nil {
		t.Fatal("empty item accepted")
	}
	milk, _ := uc.Add(ctx, "Milk")
	_, _ = uc.Add(ctx, "Bread")
	if err := uc.SetDone(ctx, milk.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := uc.SetDone(ctx, "../etc.ics", true); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal("invalid ID accepted")
	}
	list, _ := uc.List(ctx)
	if len(list) != 2 || list[0].Summary != "Bread" || !list[1].Done {
		t.Fatalf("%+v", list)
	}
	if n, err := uc.ClearCompleted(ctx); n != 1 || err != nil {
		t.Fatalf("clear: %d %v", n, err)
	}
	store.down = true
	if _, err := uc.List(ctx); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("store error not reported: %v", err)
	}
}
