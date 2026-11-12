package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/timeline/domain"
	"trisse-home/internal/modules/timeline/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "timeline", Migrations()))
	t0 := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	for i, src := range []string{"service", "security", "service"} {
		sev := domain.SeverityInfo
		if i == 2 {
			sev = domain.SeverityCritical
		}
		if err := repo.Add(ctx, domain.NewEvent(string(rune('a'+i)), t0.Add(time.Duration(i)*time.Hour), src, sev, "s", "m")); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.List(ctx, usecase.Filter{Source: "service", Limit: 10})
	if err != nil || len(list) != 2 || list[0].ID != "c" || list[0].Severity != domain.SeverityCritical {
		t.Fatalf("%+v %v", list, err)
	}
	if page, _ := repo.List(ctx, usecase.Filter{Before: t0.Add(2 * time.Hour), Limit: 10}); len(page) != 2 {
		t.Fatalf("before: %+v", page)
	}
	if d, _ := repo.RetentionDays(ctx); d != 0 {
		t.Fatal("no retention stored yet")
	}
	_ = repo.SetRetentionDays(ctx, 14)
	if d, _ := repo.RetentionDays(ctx); d != 14 {
		t.Fatal("retention not stored")
	}
	_ = repo.DeleteBefore(ctx, t0.Add(90*time.Minute))
	if all, _ := repo.List(ctx, usecase.Filter{Limit: 10}); len(all) != 1 {
		t.Fatalf("delete: %+v", all)
	}
}
