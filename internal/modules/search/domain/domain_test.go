package domain

import "testing"

func TestResolve(t *testing.T) {
	ddg, _ := ProviderByID(DefaultProviderID)
	yt, err := NewShortcut("!YT", "YouTube", "https://www.youtube.com/results?search_query={query}")
	if err != nil {
		t.Fatal(err)
	}
	shortcuts := []Shortcut{yt}
	tests := []struct {
		query, want string
	}{
		{"go generics", "https://duckduckgo.com/?q=go+generics"},
		{"!w Oldenburg", "https://duckduckgo.com/?q=%21w+Oldenburg"},
		{"!yt lofi beats", "https://www.youtube.com/results?search_query=lofi+beats"},
		{"lofi beats !yt", "https://www.youtube.com/results?search_query=lofi+beats"},
		{"https://example.com/path", "https://example.com/path"},
		{"  a&b=c  ", "https://duckduckgo.com/?q=a%26b%3Dc"},
	}
	for _, tt := range tests {
		got, err := Resolve(tt.query, ddg, shortcuts)
		if err != nil || got != tt.want {
			t.Errorf("Resolve(%q) = %q, %v; want %q", tt.query, got, err, tt.want)
		}
	}
	if _, err := Resolve("   ", ddg, nil); err == nil {
		t.Error("expected error for empty query")
	}
}

func TestNewShortcutValidation(t *testing.T) {
	tests := []struct {
		keyword, name, tmpl string
		ok                  bool
	}{
		{"gh", "GitHub", "https://github.com/search?q={query}", true},
		{"bad key", "x", "https://x.y/?q={query}", false},
		{"gh", "", "https://x.y/?q={query}", false},
		{"gh", "x", "https://x.y/?q=", false},
		{"gh", "x", "javascript:alert({query})", false},
	}
	for _, tt := range tests {
		if _, err := NewShortcut(tt.keyword, tt.name, tt.tmpl); (err == nil) != tt.ok {
			t.Errorf("NewShortcut(%q,%q,%q) err=%v want ok=%v", tt.keyword, tt.name, tt.tmpl, err, tt.ok)
		}
	}
}

func TestProviderByID(t *testing.T) {
	if _, err := ProviderByID("nope"); err == nil {
		t.Fatal("expected error")
	}
	for _, p := range Providers() {
		if err := ValidateTemplate("t", p.URL("{query}")); err != nil && p.URL("x") == "" {
			t.Errorf("provider %s broken", p.ID())
		}
	}
}
