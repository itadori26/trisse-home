package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestHistoryStore(t *testing.T) {
	ctx := context.Background()
	s := NewHistoryStore(dbtest.Open(t, "system", Migrations()))
	t0 := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	for i := range 3 {
		if err := s.Add(ctx, domain.Point{At: t0.Add(time.Duration(i) * time.Minute), CPU: float64(i)}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Since(ctx, t0.Add(time.Minute))
	if err != nil || len(got) != 2 || got[1].CPU != 2 {
		t.Fatalf("since: %+v %v", got, err)
	}
	if err := s.PruneBefore(ctx, t0.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if all, _ := s.Since(ctx, t0); len(all) != 1 {
		t.Fatalf("prune left %d points", len(all))
	}
}
