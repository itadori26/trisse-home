package gateway

import (
	"context"
	"errors"
	"strings"
	"testing"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/modules/appearance/usecase"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

func TestBuiltinThemes(t *testing.T) {
	ts := BuiltinThemes()
	if len(ts) < 6 {
		t.Fatalf("only %d valid built-in themes; a theme file is broken?", len(ts))
	}
	files, _ := builtinFiles.ReadDir("themes")
	if len(files) != len(ts) {
		t.Fatalf("%d files, %d valid themes", len(files), len(ts))
	}
	for _, th := range ts {
		if !th.BuiltIn || len(th.Tokens) < 11 {
			t.Errorf("%s: %+v", th.ID, th)
		}
	}
}

func TestDecodeTheme(t *testing.T) {
	ok := `{"id":"mine","name":"Mine","scheme":"dark","tokens":{"--color-bg":"#000","--color-surface":"#111","--color-text":"#fff","--color-accent":"#f0f"}}`
	th, err := DecodeTheme([]byte(ok))
	if err != nil || th.ID != "mine" {
		t.Fatalf("%+v %v", th, err)
	}
	if _, err := DecodeTheme([]byte(strings.Replace(ok, `"id"`, `"css":"body{}","id"`, 1))); err == nil {
		t.Fatal("unknown fields (free CSS) must be refused")
	}
	if _, err := DecodeTheme([]byte(strings.Replace(ok, "#f0f", "url(x)", 1))); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := DecodeTheme([]byte(strings.Repeat(" ", MaxThemeBytes+1))); err == nil {
		t.Fatal("size")
	}
	out, _ := EncodeTheme(th)
	if back, err := DecodeTheme(out); err != nil || back.Tokens["--color-accent"] != "#f0f" {
		t.Fatalf("round trip: %s %v", out, err)
	}
}

func TestThemeStore(t *testing.T) {
	ctx := context.Background()
	s := NewThemeStore(dbtest.Open(t, "appearance", Migrations()))
	th := BuiltinThemes()[0]
	th.BuiltIn = false
	if err := s.SaveImported(ctx, th); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Imported(ctx); len(got) != 1 || got[0].ID != th.ID || got[0].BuiltIn {
		t.Fatalf("%+v", got)
	}
	if err := s.DeleteImported(ctx, "nope"); !errors.Is(err, domain.ErrThemeNotFound) {
		t.Fatal(err)
	}
	if _, ok, _ := s.Backup(ctx); ok {
		t.Fatal("no backup yet")
	}
	_ = s.SaveBackup(ctx, usecase.StyleBackup{CSS: "body{}"})
	_ = s.SaveBackup(ctx, usecase.StyleBackup{Theme: domain.ThemeLight, CSS: "body{color:red}", AccentColor: "#112233"})
	if b, ok, _ := s.Backup(ctx); !ok || b.CSS != "body{color:red}" || b.AccentColor != "#112233" || b.Theme != domain.ThemeLight {
		t.Fatal(b)
	}
}
