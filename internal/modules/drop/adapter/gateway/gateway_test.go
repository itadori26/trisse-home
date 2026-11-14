package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/drop/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "drop", Migrations()))
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	f, _ := domain.NewFile("f", "a.pdf", "application/pdf", 4, time.Hour, now)
	if err := repo.Add(ctx, f, []byte("%PDF")); err != nil {
		t.Fatal(err)
	}
	txt, _ := domain.NewText("t", "hi", 7*24*time.Hour, now)
	_ = repo.Add(ctx, txt, nil)
	if got, err := repo.Get(ctx, "f"); err != nil || got.FileName != "a.pdf" || !got.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("%+v %v", got, err)
	}
	if data, _ := repo.File(ctx, "f"); string(data) != "%PDF" {
		t.Fatal("file")
	}
	if n, err := repo.DeleteExpired(ctx, now.Add(2*time.Hour)); err != nil || n != 1 {
		t.Fatalf("%d %v", n, err)
	}
	if _, err := repo.File(ctx, "f"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("file must go with the item")
	}
	if list, _ := repo.List(ctx); len(list) != 1 || list[0].Text != "hi" {
		t.Fatalf("%+v", list)
	}
}
