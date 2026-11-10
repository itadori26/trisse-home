package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "bookmarks", Migrations()))
	f, _ := domain.NewFolder("f1", "Dev", 0)
	if err := repo.SaveFolder(ctx, f); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	b, _ := domain.NewBookmark("b1", domain.BookmarkInput{FolderID: "f1", URL: "https://go.dev", Tags: []string{"go", "docs"}}, now)
	if err := repo.SaveBookmark(ctx, b); err != nil {
		t.Fatal(err)
	}
	got, err := repo.GetBookmark(ctx, "b1")
	if err != nil || got.FolderID() != "f1" || len(got.Tags()) != 2 || got.Tags()[0] != "docs" || !got.CreatedAt().Equal(now) {
		t.Fatalf("round trip failed: %+v %v", got, err)
	}
	_ = b.Update(domain.BookmarkInput{URL: "https://go.dev", Tags: []string{"golang"}})
	if err := repo.SaveBookmark(ctx, b); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.GetBookmark(ctx, "b1"); len(got.Tags()) != 1 || got.FolderID() != "" {
		t.Fatalf("update not applied: %+v", got)
	}
	_ = repo.SaveBookmark(ctx, func() domain.Bookmark { b.MoveToFolder("f1"); return b }())
	if err := repo.DeleteFolder(ctx, "f1"); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.GetBookmark(ctx, "b1"); got.FolderID() != "" {
		t.Fatal("bookmark not unfiled by foreign key")
	}
	if err := repo.DeleteBookmark(ctx, "b1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetBookmark(ctx, "b1"); !errors.Is(err, domain.ErrBookmarkNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
