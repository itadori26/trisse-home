package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/backupmonitor/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "backupmonitor", Migrations()))
	now := time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC)
	s, _ := domain.NewSource("s1", "NAS", domain.Schedule{Kind: "check", Cron: "0 3 * * *", Grace: 10 * time.Minute}, domain.HashToken("tok"), now)
	s.RecordRun(true, "ok", now.Add(time.Hour))
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetByToken(ctx, domain.HashToken("tok"))
	if err != nil || got.Kind != "check" || got.Cron == nil || got.Cron.String() != "0 3 * * *" || got.Grace != 10*time.Minute || !got.LastSuccessAt.Equal(now.Add(time.Hour)) || !got.LastFailureAt.IsZero() {
		t.Fatalf("round trip %+v %v", got, err)
	}
	if _, err := repo.GetByToken(ctx, domain.HashToken("nope")); !errors.Is(err, domain.ErrSourceNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	_ = repo.Delete(ctx, "s1")
	if list, _ := repo.List(ctx); len(list) != 0 {
		t.Fatal("not deleted")
	}
}
