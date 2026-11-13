package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/modules/readlater/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

var png1x1 = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\xff\xff?\x00\x05\xfe\x02\xfe\xa7\x35\x81\x84\x00\x00\x00\x00IEND\xaeB`\x82")

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "readlater", Migrations()))
	a, _ := domain.NewArticle("a", "https://x.example/post", time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC))
	a.ApplyPage("Post", "X", "e", "<p>x</p>", 10)
	a.Bytes = 100
	if err := repo.Save(ctx, a, map[int]usecase.Image{0: {Bytes: png1x1, ContentType: "image/png"}}); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, "a")
	if err != nil || got.Content != "<p>x</p>" || got.Words != 10 || got.Status != domain.StatusUnread {
		t.Fatalf("%+v %v", got, err)
	}
	if img, err := repo.Image(ctx, "a", 0); err != nil || img.ContentType != "image/png" || len(img.Bytes) != len(png1x1) {
		t.Fatalf("image: %v", err)
	}
	if n, _ := repo.TotalBytes(ctx); n != 100 {
		t.Fatal(n)
	}
	_ = repo.SetStatus(ctx, "a", domain.StatusRead)
	if got, _ := repo.Get(ctx, "a"); got.Status != domain.StatusRead {
		t.Fatal(got.Status)
	}
	_ = repo.Delete(ctx, "a")
	if _, err := repo.Image(ctx, "a", 0); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal("images must go with the article")
	}
}

func TestFetcher(t *testing.T) {
	page, _ := os.ReadFile("testdata/article.html")
	mux := http.NewServeMux()
	mux.HandleFunc("/post", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(page)
	})
	mux.HandleFunc("/img.png", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream") // wrong header, real PNG
		_, _ = w.Write(png1x1)
	})
	mux.HandleFunc("/evil.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = io.WriteString(w, `<svg onload="alert(1)"/>`)
	})
	mux.HandleFunc("/file.pdf", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = io.WriteString(w, "%PDF")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	f := NewFetcher(srv.Client())
	ctx := context.Background()
	p, err := f.Read(ctx, srv.URL+"/post")
	if err != nil || p.Title != "Warum Homeserver glücklich machen" || !strings.Contains(p.Content, "Seit drei Jahren") ||
		len(p.Images) != 1 || p.Images[0] != srv.URL+"/img/server.jpg" {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err := f.Read(ctx, srv.URL+"/file.pdf"); err == nil {
		t.Fatal("a PDF is not a web page")
	}
	if img, err := f.Fetch(ctx, srv.URL+"/img.png"); err != nil || img.ContentType != "image/png" {
		t.Fatalf("png: %+v %v", img, err)
	}
	if _, err := f.Fetch(ctx, srv.URL+"/evil.svg"); err == nil {
		t.Fatal("SVG must be refused")
	}
	if _, err := f.Fetch(ctx, srv.URL+"/missing"); err == nil {
		t.Fatal("404")
	}
}
