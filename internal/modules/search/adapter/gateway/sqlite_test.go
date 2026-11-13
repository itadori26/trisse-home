package gateway

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/modules/search/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "search", Migrations()))
	if id, err := repo.DefaultProvider(ctx); err != nil || id != "" {
		t.Fatalf("expected empty default, got %q %v", id, err)
	}
	_ = repo.SetDefaultProvider(ctx, "brave")
	_ = repo.SetDefaultProvider(ctx, "google")
	if id, _ := repo.DefaultProvider(ctx); id != "google" {
		t.Fatalf("upsert failed: %q", id)
	}
	s, _ := domain.NewShortcut("gh", "GitHub", "https://github.com/search?q={query}")
	if err := repo.AddShortcut(ctx, s); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddShortcut(ctx, s); !errors.Is(err, usecase.ErrShortcutExists) {
		t.Fatalf("expected duplicate, got %v", err)
	}
	if list, _ := repo.Shortcuts(ctx); len(list) != 1 || list[0].Name() != "GitHub" {
		t.Fatalf("unexpected shortcuts %+v", list)
	}
	if err := repo.DeleteShortcut(ctx, "gh"); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteShortcut(ctx, "gh"); !errors.Is(err, usecase.ErrShortcutNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestBarOptions(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "search", Migrations())
	repo := NewRepository(db)
	if o, err := repo.BarOptions(ctx); err != nil || o != domain.DefaultBarOptions() {
		t.Fatalf("no row yet: %+v %v", o, err)
	}
	want, _ := domain.NewBarOptions(true, domain.CapturePrefixes{Todo: "a", Note: "n", Bookmark: "lz"})
	if err := repo.SetBarOptions(ctx, want); err != nil {
		t.Fatal(err)
	}
	if o, _ := repo.BarOptions(ctx); o != want {
		t.Fatalf("round trip: %+v", o)
	}
	if id, _ := repo.DefaultProvider(ctx); id != "" {
		t.Fatalf("saving the options must not pick a provider, got %q", id)
	}
	_ = repo.SetDefaultProvider(ctx, "google")
	if o, _ := repo.BarOptions(ctx); o != want {
		t.Fatal("changing the provider must keep the options")
	}
	// A stored value that breaks today's rules falls back to the default prefixes.
	if _, err := db.ExecContext(ctx, `UPDATE search_settings SET capture_prefixes = '{"Todo":"x","Note":"x","Bookmark":"b"}'`); err != nil {
		t.Fatal(err)
	}
	if o, _ := repo.BarOptions(ctx); !o.OpenInNewTab() || o.Prefixes() != domain.DefaultCapturePrefixes {
		t.Fatalf("fallback: %+v", o)
	}
}
