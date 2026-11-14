package domain

import (
	"errors"
	"strings"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func base() map[string]string {
	return map[string]string{"--color-bg": "#2E3440", "--color-surface": "#3b4252", "--color-text": "#eceff4", "--color-accent": "#88c0d0", "--radius": "6px"}
}

func TestThemeDef(t *testing.T) {
	th, err := NewThemeDef("nord", "Nord", "Arctic Ice Studio", ThemeDark, base())
	if err != nil || th.Tokens["--color-bg"] != "#2e3440" {
		t.Fatalf("%+v %v", th, err)
	}
	css := th.CSS()
	if !strings.HasPrefix(css, "/* trisse-home theme: nord */") || !strings.Contains(css, "--color-accent: #88c0d0;") || !IsThemeCSS(css) || AppliedTheme(css) != "nord" {
		t.Fatalf("%s", css)
	}
	for name, mod := range map[string]func(map[string]string){
		"unknown token":   func(m map[string]string) { m["--font-sans"] = "#fff" },
		"url value":       func(m map[string]string) { m["--color-bg"] = "url(https://evil.example/x.png)" },
		"css injection":   func(m map[string]string) { m["--color-bg"] = "#fff; } body { display:none" },
		"expression":      func(m map[string]string) { m["--color-accent"] = "red" },
		"radius too big":  func(m map[string]string) { m["--radius"] = "999px" },
		"radius as color": func(m map[string]string) { m["--radius"] = "#fff" },
		"missing text":    func(m map[string]string) { delete(m, "--color-text") },
	} {
		tok := base()
		mod(tok)
		if _, err := NewThemeDef("x1", "X", "", ThemeLight, tok); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, id := range []string{"A", "x", "../x", "a b"} {
		if _, err := NewThemeDef(id, "X", "", ThemeLight, base()); err == nil {
			t.Errorf("id %q accepted", id)
		}
	}
	if _, err := NewThemeDef("x1", "X", "", "system", base()); err == nil {
		t.Error("scheme system")
	}
	if IsThemeCSS("body{}") || AppliedTheme("body{}") != "" {
		t.Error("own CSS is not a theme")
	}
}
