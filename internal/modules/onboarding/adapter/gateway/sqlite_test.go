package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/onboarding/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "onboarding", Migrations()))
	if s, err := repo.Load(ctx); err != nil || s.Completed() {
		t.Fatalf("expected pending state, got %v %v", s, err)
	}
	s := domain.State{}
	s.Complete(time.Date(2026, 11, 11, 8, 0, 0, 0, time.UTC))
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Load(ctx); !got.CompletedAt().Equal(s.CompletedAt()) {
		t.Fatalf("round trip: %v", got.CompletedAt())
	}
	s.Restart()
	_ = repo.Save(ctx, s)
	if got, _ := repo.Load(ctx); got.Completed() {
		t.Fatal("restart not stored")
	}
}
