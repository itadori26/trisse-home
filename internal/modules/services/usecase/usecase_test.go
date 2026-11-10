package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"trisse-home/internal/modules/services/domain"
)

type memRepo struct{ items []domain.Service }

func (m *memRepo) List(context.Context) ([]domain.Service, error) {
	out := slices.Clone(m.items)
	slices.SortFunc(out, func(a, b domain.Service) int { return a.Position - b.Position })
	return out, nil
}
func (m *memRepo) Get(_ context.Context, id string) (domain.Service, error) {
	for _, s := range m.items {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Service{}, domain.ErrServiceNotFound
}
func (m *memRepo) Create(_ context.Context, s domain.Service) error {
	m.items = append(m.items, s)
	return nil
}
func (m *memRepo) Update(_ context.Context, s domain.Service) error {
	for i := range m.items {
		if m.items[i].ID == s.ID {
			m.items[i] = s
		}
	}
	return nil
}
func (m *memRepo) Delete(_ context.Context, id string) error {
	m.items = slices.DeleteFunc(m.items, func(s domain.Service) bool { return s.ID == id })
	return nil
}

type memStatus struct {
	mu sync.Mutex
	m  map[string]domain.Status
}

func (s *memStatus) Get(id string) (domain.Status, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.m[id]
	return st, ok
}
func (s *memStatus) Put(st domain.Status) { s.mu.Lock(); s.m[st.ServiceID] = st; s.mu.Unlock() }
func (s *memStatus) Delete(id string)     { s.mu.Lock(); delete(s.m, id); s.mu.Unlock() }

type fakeChecker struct{ down map[string]bool }

func (f fakeChecker) Check(_ context.Context, url string, _ bool) (int, time.Duration, error) {
	if f.down[url] {
		return 0, 0, errors.New("connection refused")
	}
	return 200, 12 * time.Millisecond, nil
}

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprint(i.n) }

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func TestServices(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	uc := NewServices(Deps{Repo: repo, Status: &memStatus{m: map[string]domain.Status{}},
		Checker: fakeChecker{down: map[string]bool{"https://down.example": true}}, Clock: clk{}, IDs: &ids{}})

	up, err := uc.Create(ctx, ServiceInput{Name: "Up", URL: "https://up.example"})
	if err != nil || up.State != "online" || up.LatencyMs != 12 {
		t.Fatalf("create: %+v %v", up, err)
	}
	down, _ := uc.Create(ctx, ServiceInput{Name: "Down", URL: "https://down.example"})
	if down.State != "offline" || down.Error == "" {
		t.Fatalf("expected offline: %+v", down)
	}
	if _, err := uc.Create(ctx, ServiceInput{Name: "", URL: "https://x"}); err == nil {
		t.Fatal("expected validation error")
	}
	if err := uc.Reorder(ctx, []string{down.ID, up.ID}); err != nil {
		t.Fatal(err)
	}
	list, _ := uc.List(ctx)
	if list[0].ID != down.ID {
		t.Fatal("reorder not applied")
	}
	if err := uc.CheckAll(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Update(ctx, up.ID, ServiceInput{Name: "Up2", URL: "https://up.example"}); err != nil {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, "missing"); !errors.Is(err, domain.ErrServiceNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := uc.Delete(ctx, up.ID); err != nil {
		t.Fatal(err)
	}
}
