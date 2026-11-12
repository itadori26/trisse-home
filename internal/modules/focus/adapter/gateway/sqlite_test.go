package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/focus/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "focus", Migrations()))
	base := time.Date(2026, 11, 12, 10, 0, 0, 0, time.FixedZone("CET", 3600))
	for i, ago := range []int{0, 2, 9} {
		s := domain.Session{ID: string(rune('a' + i)), Minutes: 25, EndedAt: base.AddDate(0, 0, -ago)}
		if err := repo.Save(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.Since(ctx, base.AddDate(0, 0, -3))
	if err != nil || len(got) != 2 || got[0].ID != "b" || !got[1].EndedAt.Equal(base) || got[1].Minutes != 25 {
		t.Fatalf("%+v %v", got, err)
	}
	if err := repo.DeleteBefore(ctx, base.AddDate(0, 0, -5)); err != nil {
		t.Fatal(err)
	}
	if all, _ := repo.Since(ctx, time.Time{}); len(all) != 2 {
		t.Fatalf("after prune %+v", all)
	}
}
