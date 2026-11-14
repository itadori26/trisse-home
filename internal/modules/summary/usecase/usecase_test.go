package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/summary/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	s     *domain.Settings
	cache map[string]Cached
}

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }
func (m *memRepo) Clear(context.Context) error {
	m.s, m.cache = nil, map[string]Cached{}
	return nil
}
func (m *memRepo) Cached(_ context.Context, k string) (Cached, bool, error) {
	c, ok := m.cache[k]
	return c, ok, nil
}
func (m *memRepo) Store(_ context.Context, k string, c Cached) error { m.cache[k] = c; return nil }

type fakeModel struct {
	prompts []string
	err     error
}

func (f *fakeModel) Summarize(_ context.Context, _ domain.Settings, prompt string) (string, error) {
	f.prompts = append(f.prompts, prompt)
	return "  - Punkt eins\n- Punkt zwei  ", f.err
}
func (f *fakeModel) Stream(_ context.Context, _ domain.Settings, prompt string, onChunk func(string) error) error {
	f.prompts = append(f.prompts, prompt)
	for _, c := range []string{"Jellyfin ", "war offline."} {
		if err := onChunk(c); err != nil {
			return err
		}
	}
	return nil
}
func (f *fakeModel) Models(context.Context, string) ([]string, error) {
	return []string{"llama3.2:3b", "qwen2.5:7b"}, f.err
}

type fakePages struct{ url string }

func (f *fakePages) Text(_ context.Context, url string) (string, error) {
	f.url = url
	return "Der Artikel erklärt ausführlich, warum Heimserver Spaß machen und was sie kosten.", nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC) }

func TestSummary(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{cache: map[string]Cached{}}
	model := &fakeModel{}
	pages := &fakePages{}
	uc := NewSummary(Deps{Repo: repo, Model: model, Fetcher: pages, Clock: clk{}})

	if v, _ := uc.Settings(ctx); v.Configured {
		t.Fatal("not configured")
	}
	if _, err := uc.Summarize(ctx, Request{Text: strings.Repeat("Text ", 20)}); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal("hidden until configured", err)
	}
	if _, err := uc.SaveSettings(ctx, "http://ollama.lan:11434", "mistral", 0); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("model must be installed", err)
	}
	if models, err := uc.Models(ctx, "http://ollama.lan:11434"); err != nil || len(models) != 2 {
		t.Fatal(models, err)
	}
	if _, err := uc.Models(ctx, "https://api.example.com"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("only local servers", err)
	}
	v, err := uc.SaveSettings(ctx, "http://ollama.lan:11434", "llama3.2:3b", 0)
	if err != nil || !v.Configured || v.MaxInputChars != domain.DefaultMaxInputChars {
		t.Fatalf("%+v %v", v, err)
	}

	note := "Einkaufsliste für das Wochenende: Brot, Käse, Tomaten und Kaffee besorgen."
	r, err := uc.Summarize(ctx, Request{Text: note, Language: "de"})
	if err != nil || r.Text != "- Punkt eins\n- Punkt zwei" || r.Cached || !strings.Contains(model.prompts[0], "Deutsch") {
		t.Fatalf("%+v %v", r, err)
	}
	// The same note again comes from the cache; a changed note does not.
	if r, _ := uc.Summarize(ctx, Request{Text: note, Language: "de"}); !r.Cached || len(model.prompts) != 1 {
		t.Fatal("cache")
	}
	if r, _ := uc.Summarize(ctx, Request{Text: note + " Und Milch.", Language: "de"}); r.Cached || len(model.prompts) != 2 {
		t.Fatal("a new version is summarised again")
	}
	// Feed items: the page behind the link is loaded on the server.
	if r, err := uc.Summarize(ctx, Request{URL: "https://news.example/artikel?id=1", Language: "en"}); err != nil || pages.url != "https://news.example/artikel?id=1" || r.Text == "" {
		t.Fatal(r, err)
	}
	if _, err := uc.Summarize(ctx, Request{URL: "javascript:alert(1)"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	model.err = errors.New("model not loaded")
	if _, err := uc.Summarize(ctx, Request{Text: note + " Neu."}); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal(err)
	}
	_ = uc.Disable(ctx)
	if v, _ := uc.Settings(ctx); v.Configured || len(repo.cache) != 0 {
		t.Fatal("disable clears everything")
	}
}

func TestAsk(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{cache: map[string]Cached{}}
	model := &fakeModel{}
	uc := NewSummary(Deps{Repo: repo, Model: model, Fetcher: &fakePages{}, Clock: clk{}})
	var got strings.Builder
	collect := func(c string) error { got.WriteString(c); return nil }
	if err := uc.Ask(ctx, "? Warum war Jellyfin offline?", "service jellyfin: offline", "de", collect); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal("hidden until configured", err)
	}
	_, _ = uc.SaveSettings(ctx, "http://ollama:11434", "llama3.2:3b", 0)
	if err := uc.Ask(ctx, "?", "", "de", collect); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	if err := uc.Ask(ctx, "? Warum war Jellyfin offline?", "service jellyfin: offline", "de", collect); err != nil || got.String() != "Jellyfin war offline." {
		t.Fatal(got.String(), err)
	}
	if p := model.prompts[0]; !strings.Contains(p, "service jellyfin: offline") || !strings.HasSuffix(p, "Frage: Warum war Jellyfin offline?") {
		t.Fatal(p)
	}
	stop := errors.New("client gone")
	if err := uc.Ask(ctx, "? noch eine Frage", "", "en", func(string) error { return stop }); !errors.Is(err, stop) {
		t.Fatal("a cancelled stream stops", err)
	}
}
