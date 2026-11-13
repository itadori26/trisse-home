package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	provider  string
	shortcuts []domain.Shortcut
	bar       *domain.BarOptions
}

func (m *memRepo) BarOptions(context.Context) (domain.BarOptions, error) {
	if m.bar == nil {
		return domain.DefaultBarOptions(), nil
	}
	return *m.bar, nil
}
func (m *memRepo) SetBarOptions(_ context.Context, o domain.BarOptions) error { m.bar = &o; return nil }

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

func TestSearchFieldOptions(t *testing.T) {
	repo := &memRepo{}
	uc := NewSearcher(repo)
	res, err := uc.Settings(context.Background())
	if err != nil || res.Bar != (BarView{TodoPrefix: "t", NotePrefix: "n", BookmarkPrefix: "b"}) {
		t.Fatalf("defaults: %+v %v", res.Bar, err)
	}
	res, err = uc.SetBarOptions(context.Background(), SetBarRequest{OpenInNewTab: true, TodoPrefix: "A", NotePrefix: "n", BookmarkPrefix: "lz"})
	if err != nil || res.Bar != (BarView{OpenInNewTab: true, TodoPrefix: "a", NotePrefix: "n", BookmarkPrefix: "lz"}) {
		t.Fatalf("saved: %+v %v", res.Bar, err)
	}
	if _, err := uc.SetBarOptions(context.Background(), SetBarRequest{TodoPrefix: "x", NotePrefix: "x", BookmarkPrefix: "b"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("duplicate prefixes must be refused, got %v", err)
	}
	if repo.bar.Prefixes().Todo != "a" {
		t.Fatal("a refused request must not change the stored options")
	}
}
