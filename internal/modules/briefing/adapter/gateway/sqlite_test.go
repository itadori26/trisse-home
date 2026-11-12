package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/briefing/domain"
	"trisse-home/internal/modules/briefing/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestSettingsRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "briefing", Migrations()))
	if _, ok, err := repo.Load(ctx); ok || err != nil {
		t.Fatalf("empty table: %v %v", ok, err)
	}
	push, _ := domain.NewPushSchedule(true, "06:45", []int{1, 5}, false)
	in := usecase.Settings{
		Layout:   domain.Layout{{Kind: domain.SectionTodos, Enabled: true}, {Kind: domain.SectionWaste}},
		Push:     push,
		Lat:      53.07,
		Lon:      8.8,
		LastSent: time.Date(2026, 11, 12, 6, 45, 0, 0, time.UTC),
	}
	if err := repo.Save(ctx, in); err != nil {
		t.Fatal(err)
	}
	got, ok, err := repo.Load(ctx)
	if err != nil || !ok {
		t.Fatalf("%v %v", ok, err)
	}
	if len(got.Layout) != 2 || got.Layout[1] != (domain.LayoutEntry{Kind: domain.SectionWaste}) ||
		got.Push.At() != "06:45" || len(got.Push.Days) != 2 || got.Push.Days[1] != time.Friday || got.Push.SkipEmpty ||
		got.Lat != 53.07 || got.Lon != 8.8 || !got.LastSent.Equal(in.LastSent) {
		t.Fatalf("got %+v", got)
	}
	in.LastSent = time.Time{}
	if err := repo.Save(ctx, in); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := repo.Load(ctx); !got.LastSent.IsZero() {
		t.Fatal("last sent not cleared")
	}
}
