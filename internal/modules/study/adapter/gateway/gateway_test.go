package gateway

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/study/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "study", Migrations()))
	if v, err := repo.TotalECTS(ctx); err != nil || v != domain.DefaultTotalECTS {
		t.Fatalf("%v %v", v, err)
	}
	_ = repo.SetTotalECTS(ctx, 210)
	if v, _ := repo.TotalECTS(ctx); v != 210 {
		t.Fatal(v)
	}
	grade := 1.7
	in := domain.Module{ID: "a", Name: "Analysis", ECTS: 7.5, Semester: 1, Status: domain.StatusPassed, Grade: &grade}
	if err := repo.Save(ctx, in); err != nil {
		t.Fatal(err)
	}
	_ = repo.Save(ctx, domain.Module{ID: "b", Name: "Praktikum", ECTS: 15, Status: domain.StatusOpen})
	got, err := repo.Get(ctx, "a")
	if err != nil || got.ECTS != 7.5 || *got.Grade != 1.7 || got.Status != domain.StatusPassed {
		t.Fatalf("%+v %v", got, err)
	}
	if b, _ := repo.Get(ctx, "b"); b.Grade != nil || b.Semester != 0 {
		t.Fatalf("%+v", b)
	}
	in.Grade, in.Status = nil, domain.StatusInProgress
	_ = repo.Save(ctx, in)
	if got, _ := repo.Get(ctx, "a"); got.Grade != nil || got.Status != domain.StatusInProgress {
		t.Fatalf("update: %+v", got)
	}
	if list, _ := repo.List(ctx); len(list) != 2 {
		t.Fatal(list)
	}
	if err := repo.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, "a"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, "a"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
}
