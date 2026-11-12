package gateway

import (
	"context"
	"testing"

	"trisse-home/internal/modules/system/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestAggregateStore(t *testing.T) {
	ctx := context.Background()
	s := NewAggregateStore(dbtest.Open(t, "system", Migrations()))
	if _, ok, err := s.Get(ctx, "/", "2026-11-01"); ok || err != nil {
		t.Fatalf("empty: %v %v", ok, err)
	}
	for _, a := range []domain.UsageAggregate{
		{Mount: "/", Day: "2026-11-01", MinUsed: 1, MaxUsed: 3, Last: 2, Total: 4 << 40},
		{Mount: "/", Day: "2026-11-02", MinUsed: 2, MaxUsed: 4, Last: 4, Total: 4 << 40},
		{Mount: "/data", Day: "2026-11-02", Last: 9, Total: 10},
	} {
		if err := s.Put(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	_ = s.Put(ctx, domain.UsageAggregate{Mount: "/", Day: "2026-11-02", MinUsed: 2, MaxUsed: 5, Last: 5, Total: 4 << 40})
	got, ok, _ := s.Get(ctx, "/", "2026-11-02")
	if !ok || got.MaxUsed != 5 || got.Total != 4<<40 {
		t.Fatalf("upsert: %+v", got)
	}
	list, err := s.Since(ctx, "2026-11-02")
	if err != nil || len(list) != 2 || list[0].Mount != "/" || list[1].Mount != "/data" {
		t.Fatalf("since: %+v %v", list, err)
	}
	_ = s.PruneBefore(ctx, "2026-11-02")
	if all, _ := s.Since(ctx, "2000-01-01"); len(all) != 2 {
		t.Fatalf("prune: %+v", all)
	}
	if _, ok, _ := s.ForecastSettings(ctx); ok {
		t.Fatal("no settings yet")
	}
	_ = s.SaveForecastSettings(ctx, domain.ForecastSettings{WindowDays: 30, AlertDays: 7})
	if fs, ok, _ := s.ForecastSettings(ctx); !ok || fs.WindowDays != 30 || fs.AlertDays != 7 {
		t.Fatalf("settings: %+v", fs)
	}
}
