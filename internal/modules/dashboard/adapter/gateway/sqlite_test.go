package gateway

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/dashboard/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepositoryRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "dashboard", Migrations()))

	tab, _ := domain.NewTab("t1", "Home", 0)
	if err := repo.CreateTab(ctx, tab); err != nil {
		t.Fatal(err)
	}
	rect, _ := domain.NewRect(1, 2, 3, 4)
	w, _ := domain.NewWidget("w1", "t1", "clock", rect, `{"a":1}`)
	if err := repo.CreateWidget(ctx, w); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetWidget(ctx, "w1")
	if err != nil || got.Rect() != rect || got.Config() != `{"a":1}` {
		t.Fatalf("round trip failed: %+v %v", got, err)
	}
	moved, _ := domain.NewRect(0, 0, 12, 1)
	got.Place(moved)
	if err := repo.SaveWidgetRects(ctx, []domain.Widget{got}); err != nil {
		t.Fatal(err)
	}
	if again, _ := repo.GetWidget(ctx, "w1"); again.Rect() != moved {
		t.Fatalf("rect not saved: %+v", again.Rect())
	}
	if err := repo.DeleteTab(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetWidget(ctx, "w1"); !errors.Is(err, domain.ErrWidgetNotFound) {
		t.Fatalf("widget not removed with tab: %v", err)
	}
	if _, err := repo.GetTab(ctx, "t1"); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("expected tab not found, got %v", err)
	}
}
