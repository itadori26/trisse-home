package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/summary/domain"
	"trisse-home/internal/modules/summary/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	r := NewRepository(dbtest.Open(t, "summary", Migrations()))
	if _, err := r.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatal(err)
	}
	s := domain.Settings{URL: "http://ollama:11434", Model: "m", MaxInputChars: 5000}
	_ = r.Save(ctx, s)
	if got, _ := r.Load(ctx); got != s {
		t.Fatal(got)
	}
	start := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	for i := range MaxCached + 5 {
		_ = r.Store(ctx, fmt.Sprint("k", i), usecase.Cached{Text: "s", Model: "m", CreatedAt: start.Add(time.Duration(i) * time.Second)})
	}
	if _, ok, _ := r.Cached(ctx, "k0"); ok {
		t.Fatal("the oldest summaries are pruned")
	}
	if c, ok, _ := r.Cached(ctx, fmt.Sprint("k", MaxCached+4)); !ok || c.Text != "s" {
		t.Fatal("newest kept")
	}
	_ = r.Clear(ctx)
	if _, ok, _ := r.Cached(ctx, fmt.Sprint("k", MaxCached+4)); ok {
		t.Fatal("clear")
	}
}

func TestOllama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models":[{"name":"qwen2.5:7b"},{"name":"llama3.2:3b"}]}`))
		case "/api/generate":
			var body struct {
				Model  string `json:"model"`
				Prompt string `json:"prompt"`
				Stream bool   `json:"stream"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Model == "missing" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"model 'missing' not found"}`))
				return
			}
			if body.Stream || !strings.HasSuffix(body.Prompt, "TEXT") {
				t.Errorf("request %+v", body)
			}
			_, _ = w.Write([]byte(`{"response":"- kurz","done":true}`))
		}
	}))
	defer srv.Close()
	o := NewOllama(srv.Client())
	if m, err := o.Models(context.Background(), srv.URL); err != nil || strings.Join(m, ",") != "llama3.2:3b,qwen2.5:7b" {
		t.Fatal(m, err)
	}
	if s, err := o.Summarize(context.Background(), domain.Settings{URL: srv.URL, Model: "llama3.2:3b"}, "prompt TEXT"); err != nil || s != "- kurz" {
		t.Fatal(s, err)
	}
	if _, err := o.Summarize(context.Background(), domain.Settings{URL: srv.URL, Model: "missing"}, "TEXT"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatal(err)
	}
}

func TestArticle(t *testing.T) {
	page := `<html><head><title>T</title><style>.x{}</style></head><body><nav>Menü Start Kontakt</nav>
<article><h1>Heimserver</h1><p>Ein Heimserver spart Geld &amp; Nerven.</p><script>track()</script><p>Zweiter Absatz.</p></article>
<footer>Impressum</footer></body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pdf" {
			w.Header().Set("Content-Type", "application/pdf")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}))
	defer srv.Close()
	text, err := NewArticle(srv.Client()).Text(context.Background(), srv.URL)
	if err != nil || text != "Heimserver\nEin Heimserver spart Geld & Nerven.\nZweiter Absatz." {
		t.Fatalf("%q %v", text, err)
	}
	if _, err := NewArticle(srv.Client()).Text(context.Background(), srv.URL+"/pdf"); err == nil {
		t.Fatal("not a web page")
	}
	if got := ExtractText(`<body><main><p>Nur main</p></main><aside>Werbung</aside></body>`); got != "Nur main" {
		t.Fatalf("%q", got)
	}
}
