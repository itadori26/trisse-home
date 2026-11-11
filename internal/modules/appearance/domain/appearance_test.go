package domain

import (
	"strings"
	"testing"
)

func TestNewAppearance(t *testing.T) {
	tests := []struct {
		name            string
		theme           Theme
		accent, bg, css string
		wantErr         bool
	}{
		{"defaults", ThemeSystem, "", "", "", false},
		{"full", ThemeDark, "#AABBCC", "https://example.com/bg.jpg", "body{}", false},
		{"bad theme", "pink", "", "", "", true},
		{"bad color", ThemeLight, "red", "", "", true},
		{"javascript url", ThemeLight, "", "javascript:alert(1)", "", true},
		{"css injection", ThemeLight, "", "https://x.y/a.png);background:red", "", true},
		{"huge css", ThemeLight, "", "", strings.Repeat("a", MaxCustomCSSBytes+1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAppearance(tt.theme, tt.accent, tt.bg, tt.css)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.accent != "" && a.AccentColor() != strings.ToLower(tt.accent) {
				t.Fatal("accent color not normalised")
			}
		})
	}
}

func TestLanguage(t *testing.T) {
	a := Default()
	if err := a.SetLanguage(" DE "); err != nil || a.Language() != "de" {
		t.Fatalf("%q %v", a.Language(), err)
	}
	if err := a.SetLanguage("fr"); err == nil {
		t.Fatal("unsupported language accepted")
	}
}
