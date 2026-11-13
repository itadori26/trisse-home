package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/countdown/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "countdown", Migrations()))
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timed, _ := domain.New("1", domain.Input{Title: "Exam", Emoji: "📚", Date: "2026-12-24", Time: "09:30"}, created)
	whole, _ := domain.New("2", domain.Input{Title: "Holiday", Date: "2026-12-01"}, created)
	for _, c := range []domain.Countdown{timed, whole} {
		if err := repo.Save(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.Get(ctx, "1")
	if err != nil || got.Title != "Exam" || got.Emoji != "📚" || got.Date() != "2026-12-24" || got.Time() != "09:30" || !got.CreatedAt.Equal(created) {
		t.Fatalf("%+v %v", got, err)
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 || list[0].ID != "2" || !list[0].AllDay() {
		t.Fatalf("%+v %v", list, err)
	}
	upd, _ := timed.Update(domain.Input{Title: "Exam 2", Date: "2026-12-25"})
	if err := repo.Save(ctx, upd); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(ctx, "1"); got.Title != "Exam 2" || !got.AllDay() || got.Emoji != "" {
		t.Fatalf("after update %+v", got)
	}
	_ = repo.Delete(ctx, "1")
	if _, err := repo.Get(ctx, "1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
