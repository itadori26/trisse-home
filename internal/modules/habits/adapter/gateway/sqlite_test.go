package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/habits/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "habits", Migrations()))
	created := time.Date(2026, 11, 1, 8, 0, 0, 0, time.UTC)
	weekdays, _ := domain.NewSchedule([]int{5, 1})
	gym, _ := domain.New("g", "Gym", "🏋", weekdays, 1, created)
	read, _ := domain.New("r", "Read", "", domain.Schedule{}, 0, created)
	for _, h := range []domain.Habit{gym, read} {
		if err := repo.Save(ctx, h); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 || list[0].ID != "r" || list[1].Schedule.Days()[0] != 1 || !list[0].Schedule.Daily() {
		t.Fatalf("%+v %v", list, err)
	}

	day := func(d int) domain.Day { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	for _, d := range []int{10, 11, 12} {
		if err := repo.SetCompletion(ctx, "r", day(d), true); err != nil {
			t.Fatal(err)
		}
	}
	_ = repo.SetCompletion(ctx, "r", day(12), true) // idempotent
	_ = repo.SetCompletion(ctx, "r", day(11), false)
	done, err := repo.Completions(ctx, day(11))
	if err != nil || len(done["r"]) != 1 || !done["r"][day(12)] {
		t.Fatalf("completions %+v %v", done, err)
	}

	// Deleting a habit removes its completions (foreign key cascade).
	if err := repo.Delete(ctx, "r"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, "r"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("get after delete: %v", err)
	}
	if all, _ := repo.Completions(ctx, time.Time{}); len(all) != 0 {
		t.Fatalf("completions left: %+v", all)
	}
}
