package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/platform/httpclient"
	"trisse-home/internal/shared/kernel"
)

type reverseCipher struct{}

func (reverseCipher) Encrypt(s string) (string, error) { return "enc:" + reverse(s), nil }
func (reverseCipher) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "enc:") {
		return "", errors.New("bad ciphertext")
	}
	return reverse(strings.TrimPrefix(s, "enc:")), nil
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func endpoint(id string) domain.Endpoint {
	warn := 70.0
	return domain.Endpoint{ID: id, Name: "NAS " + id, URL: "http://nas/api", HeaderName: "X-Api-Key", HeaderValue: "s3cret",
		Interval: 2 * time.Minute, Extractions: []domain.Extraction{{Label: "Temp", Path: "$.temp", Unit: "°C", Warn: &warn, Below: true}}}
}

func TestRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "apiwidget", Migrations())
	repo := NewRepository(db, reverseCipher{})
	for _, id := range []string{"b", "a"} {
		if err := repo.Save(ctx, endpoint(id)); err != nil {
			t.Fatal(err)
		}
	}
	var stored string
	_ = db.QueryRow(`SELECT header_value_encrypted FROM apiwidget_endpoints WHERE id = 'a'`).Scan(&stored)
	if strings.Contains(stored, "s3cret") {
		t.Fatal("header value stored in plain text")
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 || list[0].ID != "b" {
		t.Fatalf("order: %+v %v", list, err)
	}
	got, err := repo.Get(ctx, "a")
	want := endpoint("a")
	if err != nil || got.HeaderValue != "s3cret" || got.Interval != want.Interval || *got.Extractions[0].Warn != 70 || !got.Extractions[0].Below {
		t.Fatalf("%+v %v", got, err)
	}
	got.Name = "Renamed"
	_ = repo.Save(ctx, got)
	if list, _ := repo.List(ctx); list[1].Name != "Renamed" || list[0].ID != "b" {
		t.Fatalf("update keeps the position: %+v", list)
	}
	if err := repo.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, "a"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, "a"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
}

// An import without the secret (excluded from exports) drops the header.
func TestImportedEndpointWithoutSecret(t *testing.T) {
	db := dbtest.Open(t, "apiwidget", Migrations())
	if _, err := db.Exec(`INSERT INTO apiwidget_endpoints (id, name, url, header_name, interval_seconds, extractions)
		VALUES ('x', 'n', 'http://nas/api', 'X-Api-Key', 60, '[]')`); err != nil {
		t.Fatal(err)
	}
	e, err := NewRepository(db, reverseCipher{}).Get(context.Background(), "x")
	if err != nil || e.HasHeader() {
		t.Fatalf("%+v %v", e, err)
	}
	if len(TransferTables[0].Exclude) != 1 || TransferTables[0].Exclude[0] != "header_value_encrypted" {
		t.Fatal("the secret must not be exported")
	}
}

func TestFetcher(t *testing.T) {
	var gotKey string
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-Api-Key")
		_, _ = io.WriteString(w, `{"temp": 21.5, "big": 12345678901234567890}`)
	})
	mux.HandleFunc("/html", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "<html>") })
	mux.HandleFunc("/denied", func(w http.ResponseWriter, _ *http.Request) { http.Error(w, `{"error":"no"}`, http.StatusUnauthorized) })
	mux.HandleFunc("/huge", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `"`+strings.Repeat("x", MaxResponseBytes)+`"`)
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/ok", http.StatusFound) })
	srv := httptest.NewServer(mux)
	defer srv.Close()
	f := Fetcher{Private: httpclient.New(httpclient.WithPrivateOnly()), Public: httpclient.New(httpclient.WithIntegrationGuard())}
	at := func(path string) domain.Endpoint {
		e := endpoint("a")
		e.URL = srv.URL + path
		return e
	}
	ctx := context.Background()

	res, err := f.Fetch(ctx, at("/ok"))
	if err != nil || gotKey != "s3cret" || string(res.Raw) == "" {
		t.Fatalf("%+v %v key=%q", res, err, gotKey)
	}
	big, _ := JSONPath{}.Evaluate(res.Doc, "$.big")
	if (domain.Extraction{}).Evaluate(big).Text != "12345678901234567890" {
		t.Fatalf("big numbers must keep their digits: %v", big)
	}
	for path, want := range map[string]string{
		"/html":     "not JSON",
		"/denied":   "401",
		"/huge":     "larger than 1 MB",
		"/redirect": "302",
	} {
		_, err := f.Fetch(ctx, at(path))
		var public *kernel.PublicError
		if !errors.As(err, &public) || !strings.Contains(public.Message, want) {
			t.Errorf("%s: %v, want %q", path, err, want)
		}
	}
	// A public address needs the explicit opt-in; the guard refuses before connecting.
	pub := endpoint("p")
	pub.URL = "http://93.184.215.14/api"
	if _, err := f.Fetch(ctx, pub); err == nil || !strings.Contains(err.Error(), "Allow public addresses") {
		t.Fatalf("public address: %v", err)
	}
}
