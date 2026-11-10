package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/search/domain"
)

type memRepo struct {
	provider  string
	shortcuts []domain.Shortcut
}

func (m *memRepo) DefaultProvider(context.Context) (string, error)       { return m.provider, nil }
func (m *memRepo) SetDefaultProvider(_ context.Context, id string) error { m.provider = id; return nil }
func (m *memRepo) Shortcuts(context.Context) ([]domain.Shortcut, error)  { return m.shortcuts, nil }
func (m *memRepo) AddShortcut(_ context.Context, s domain.Shortcut) error {
	for _, e := range m.shortcuts {
		if e.Keyword() == s.Keyword() {
			return ErrShortcutExists
		}
	}
	m.shortcuts = append(m.shortcuts, s)
	return nil
}
func (m *memRepo) DeleteShortcut(_ context.Context, k string) error {
	for i, s := range m.shortcuts {
		if s.Keyword() == k {
			m.shortcuts = append(m.shortcuts[:i], m.shortcuts[i+1:]...)
			return nil
		}
	}
	return ErrShortcutNotFound
}

func TestSearcher(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	uc := NewSearcher(repo)
	if got, _ := uc.Resolve(ctx, "hello"); got != "https://duckduckgo.com/?q=hello" {
		t.Fatalf("default provider not DuckDuckGo: %s", got)
	}
	if _, err := uc.SetDefaultProvider(ctx, "altavista"); err == nil {
		t.Fatal("expected unknown provider error")
	}
	if _, err := uc.SetDefaultProvider(ctx, "startpage"); err != nil {
		t.Fatal(err)
	}
	if got, _ := uc.Resolve(ctx, "hello"); got != "https://www.startpage.com/do/search?q=hello" {
		t.Fatalf("provider not switched: %s", got)
	}
	req := AddShortcutRequest{Keyword: "gh", Name: "GitHub", URLTemplate: "https://github.com/search?q={query}"}
	if _, err := uc.AddShortcut(ctx, req); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.AddShortcut(ctx, req); !errors.Is(err, ErrShortcutExists) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
	settings, _ := uc.Settings(ctx)
	if len(settings.Shortcuts) != 1 || len(settings.Providers) < 3 {
		t.Fatalf("unexpected settings %+v", settings)
	}
	if _, err := uc.DeleteShortcut(ctx, "nope"); !errors.Is(err, ErrShortcutNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
