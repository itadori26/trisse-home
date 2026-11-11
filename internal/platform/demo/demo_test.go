package demo

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func get(t *testing.T, url string) (int, string) {
	t.Helper()
	c := &http.Client{Transport: Transport{Now: func() time.Time { return time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC) }}}
	res, err := c.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(b)
}

func TestTransport(t *testing.T) {
	code, body := get(t, "https://api.open-meteo.com/v1/forecast?latitude=1")
	var f struct {
		Daily struct{ Time []string } `json:"daily"`
	}
	if code != 200 || json.Unmarshal([]byte(body), &f) != nil || f.Daily.Time[0] != "2026-11-11" {
		t.Fatalf("forecast: %d %s", code, body)
	}
	if code, body := get(t, "http://docker-socket-proxy:2375/containers/json?all=1"); code != 200 || len(body) < 50 {
		t.Fatalf("docker: %d", code)
	}
	if code, _ := get(t, "https://go.dev/blog/feed.atom"); code != 200 {
		t.Fatalf("feed: %d", code)
	}
	if code, _ := get(t, "https://api.github.com/repos/x/y/releases/latest"); code != http.StatusServiceUnavailable {
		t.Fatalf("other hosts must be refused, got %d", code)
	}
}

func TestReadOnlySecrets(t *testing.T) {
	if _, err := (ReadOnlySecrets{}).Encrypt("x"); err == nil {
		t.Fatal("encrypt must fail")
	}
}

func TestCheckDatabase(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if err := CheckDatabase(ctx, db); err != nil {
		t.Fatalf("empty database: %v", err)
	}
	_, _ = db.Exec(`CREATE TABLE auth_accounts (id TEXT)`)
	if err := CheckDatabase(ctx, db); err != nil {
		t.Fatalf("no account yet: %v", err)
	}
	_, _ = db.Exec(`INSERT INTO auth_accounts VALUES ('owner')`)
	if err := CheckDatabase(ctx, db); err == nil {
		t.Fatal("production database accepted")
	}
	_ = MarkDatabase(ctx, db)
	if err := CheckDatabase(ctx, db); err != nil {
		t.Fatalf("demo database refused: %v", err)
	}
}

func TestBlock(t *testing.T) {
	h := Block(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	for req, want := range map[*http.Request]int{
		httptest.NewRequest(http.MethodPut, "/api/auth/password", nil):    http.StatusForbidden,
		httptest.NewRequest(http.MethodPost, "/api/auth/totp/setup", nil): http.StatusForbidden,
		httptest.NewRequest(http.MethodGet, "/api/auth/sessions", nil):    http.StatusTeapot,
		httptest.NewRequest(http.MethodPost, "/api/dashboard/tabs", nil):  http.StatusTeapot,
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("%s %s: %d, want %d", req.Method, req.URL.Path, rec.Code, want)
		}
	}
}
