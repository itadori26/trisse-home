package usecase

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"trisse-home/internal/modules/study/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	m     map[string]domain.Module
	total float64
}

func (r *memRepo) List(context.Context) ([]domain.Module, error) {
	out := []domain.Module{}
	for _, m := range r.m {
		out = append(out, m)
	}
	return out, nil
}
func (r *memRepo) Get(_ context.Context, id string) (domain.Module, error) {
	m, ok := r.m[id]
	if !ok {
		return domain.Module{}, domain.ErrNotFound
	}
	return m, nil
}
func (r *memRepo) Save(_ context.Context, m domain.Module) error { r.m[m.ID] = m; return nil }
func (r *memRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.m[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.m, id)
	return nil
}
func (r *memRepo) TotalECTS(context.Context) (float64, error) {
	if r.total == 0 {
		return domain.DefaultTotalECTS, nil
	}
	return r.total, nil
}
func (r *memRepo) SetTotalECTS(_ context.Context, v float64) error { r.total = v; return nil }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "m" + strconv.Itoa(i.n) }

func TestStudy(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{m: map[string]domain.Module{}}
	uc := New(repo, &ids{})
	grade := 2.0
	a, err := uc.Create(ctx, domain.ModuleInput{Name: "Datenbanken", ECTS: 6, Semester: 2, Status: domain.StatusPassed, Grade: &grade})
	if err != nil {
		t.Fatal(err)
	}
	_, _ = uc.Create(ctx, domain.ModuleInput{Name: "Analysis", ECTS: 9, Semester: 1, Status: domain.StatusInProgress})
	p, ms, err := uc.Progress(ctx)
	if err != nil || p.TotalECTS != 180 || p.PassedECTS != 6 || p.InProgressECTS != 9 || *p.Average != 2 || ms[0].Name != "Analysis" {
		t.Fatalf("%+v %+v %v", p, ms, err)
	}
	if _, err := uc.SetTotalECTS(ctx, 0); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("total 0")
	}
	if v, _ := uc.SetTotalECTS(ctx, 210); v != 210 {
		t.Fatal("total")
	}
	if _, err := uc.Update(ctx, a.ID, domain.ModuleInput{Name: "Datenbanken", ECTS: 6, Semester: 2, Status: domain.StatusOpen}); err != nil {
		t.Fatal(err)
	}
	if p, _, _ := uc.Progress(ctx); p.Average != nil || p.TotalECTS != 210 || p.OpenECTS != 6 {
		t.Fatalf("%+v", p)
	}
	if _, err := uc.Update(ctx, "nope", domain.ModuleInput{Name: "x", ECTS: 1, Status: domain.StatusOpen}); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, a.ID); err != nil || len(repo.m) != 1 {
		t.Fatal(err)
	}
}
