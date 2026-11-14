package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/space/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

func TestLaunchLibrary(t *testing.T) {
	var query string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		http.ServeFile(w, r, "testdata/ll2-upcoming.json")
	}))
	defer srv.Close()
	ls, err := NewLaunchLibrary(srv.Client(), srv.URL+"/").Upcoming(context.Background(), 10)
	if err != nil || len(ls) != 3 || !strings.Contains(query, "limit=10") {
		t.Fatalf("%v %v %s", ls, err, query)
	}
	l := ls[0]
	if l.Rocket != "Starship" || l.Provider != "SpaceX" || l.Status != domain.StatusGo || l.Location != "SpaceX Starbase, TX, USA" ||
		!strings.HasPrefix(l.Mission, "Starlink Group 31-2") || l.Net.Format("2006-01-02T15:04") != "2026-11-13T22:45" ||
		len([]rune(l.Description)) > domain.MaxDescription || !l.WebcastLive {
		t.Fatalf("%+v", l)
	}
}

type hostRewrite struct{ target string }

// RoundTrip sends every request to the test server, keeping the path.
func (h hostRewrite) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.URL.Scheme, r2.URL.Host = "http", strings.TrimPrefix(h.target, "http://")
	return http.DefaultTransport.RoundTrip(r2)
}

func TestAPOD(t *testing.T) {
	var key string
	mux := http.NewServeMux()
	mux.HandleFunc("/apod", func(w http.ResponseWriter, r *http.Request) {
		key = r.URL.Query().Get("api_key")
		http.ServeFile(w, r, "testdata/apod.json")
	})
	mux.HandleFunc("/apod/image/2611/Pleiades_Doe_1024.jpg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = io.WriteString(w, "\xff\xd8jpeg")
	})
	mux.HandleFunc("/evil.svg", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = io.WriteString(w, "<svg/>")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	a := NewAPOD(srv.Client(), srv.URL+"/apod")
	p, err := a.Today(context.Background(), "")
	if err != nil || key != "DEMO_KEY" || p.Title != "The Pleiades Star Cluster" || p.Copyright != "Jane Doe" || !p.IsImage() {
		t.Fatalf("%+v %v %q", p, err, key)
	}
	_, _ = a.Today(context.Background(), "mine")
	if key != "mine" {
		t.Fatal(key)
	}
	// Images are only proxied from NASA hosts over https.
	proxy := NewAPOD(&http.Client{Transport: hostRewrite{target: srv.URL}}, "")
	img, err := proxy.Image(context.Background(), p.URL)
	if err != nil || img.ContentType != "image/jpeg" || string(img.Bytes) != "\xff\xd8jpeg" {
		t.Fatalf("%+v %v", img, err)
	}
	for _, u := range []string{"http://apod.nasa.gov/a.jpg", "https://evil.example/a.jpg", "https://nasa.gov.evil.example/a.jpg", "https://apod.nasa.gov/evil.svg"} {
		if _, err := proxy.Image(context.Background(), u); err == nil {
			t.Errorf("%s must be refused", u)
		} else if u != "https://apod.nasa.gov/evil.svg" && !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", u, err)
		}
	}
}

type reverseCipher struct{}

func (reverseCipher) Encrypt(s string) (string, error) { return "enc:" + s + "!", nil }
func (reverseCipher) Decrypt(s string) (string, error) {
	return strings.TrimSuffix(strings.TrimPrefix(s, "enc:"), "!"), nil
}

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "space", Migrations())
	repo := NewSettingsRepository(db, reverseCipher{})
	if s, err := repo.Load(ctx); err != nil || s.Launches || s.Picture {
		t.Fatalf("off by default: %+v %v", s, err)
	}
	_ = repo.Save(ctx, domain.Settings{Launches: true, Picture: true, NASAKey: "k3y"})
	var enc string
	_ = db.QueryRow(`SELECT nasa_key_encrypted FROM space_settings`).Scan(&enc)
	if enc != "enc:k3y!" {
		t.Fatalf("stored: %q", enc)
	}
	if s, _ := repo.Load(ctx); !s.Launches || !s.Picture || s.NASAKey != "k3y" {
		t.Fatalf("%+v", s)
	}
}
