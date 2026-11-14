package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"trisse-home/internal/modules/adguard/domain"
	"trisse-home/internal/platform/database/dbtest"
)

// reverseCipher is a reversible stand-in for AES-GCM in tests.
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

func TestSettingsRepositoryEncryptsPassword(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "adguard", Migrations())
	repo := NewSettingsRepository(db, reverseCipher{})
	if _, err := repo.Load(ctx, domain.ProviderAdGuard); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s, _ := domain.NewSettings(domain.ProviderAdGuard, "http://ag:3000", "admin", "s3cret")
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT password_encrypted FROM adguard_connections`).Scan(&stored)
	if stored == "s3cret" || strings.Contains(stored, "s3cret") {
		t.Fatalf("password stored in plain text: %q", stored)
	}
	if got, err := repo.Load(ctx, domain.ProviderAdGuard); err != nil || got != s {
		t.Fatalf("round trip: %+v %v", got, err)
	}
	pi, _ := domain.NewSettings(domain.ProviderPiHole, "http://pi.hole", "", "app-pw")
	if err := repo.Save(ctx, pi); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.Load(ctx, domain.ProviderPiHole); err != nil || got != pi {
		t.Fatalf("providers are stored side by side: %+v %v", got, err)
	}
	if got, _ := repo.Load(ctx, domain.ProviderAdGuard); got != s {
		t.Fatalf("saving Pi-hole must not touch AdGuard: %+v", got)
	}
}

// The migration moves an existing AdGuard connection to the new table.
func TestMigrationKeepsAdGuardConnection(t *testing.T) {
	ctx := context.Background()
	first, err := fs.ReadFile(Migrations(), "0001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	db := dbtest.Open(t, "adguard", fstest.MapFS{"0001_init.sql": {Data: first}})
	if _, err := db.ExecContext(ctx, `INSERT INTO adguard_settings (id, url, username, password_encrypted) VALUES (1, 'http://ag', 'admin', 'enc:wp')`); err != nil {
		t.Fatal(err)
	}
	second, err := fs.ReadFile(Migrations(), "0002_providers.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(second)); err != nil {
		t.Fatal(err)
	}
	got, err := NewSettingsRepository(db, reverseCipher{}).Load(ctx, domain.ProviderAdGuard)
	if err != nil || got.URL != "http://ag" || got.Password != "pw" {
		t.Fatalf("migrated: %+v %v", got, err)
	}
}

func TestPiHole(t *testing.T) {
	logins := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth" {
			var body struct{ Password string }
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Password != "app-pw" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"session":{"valid":false}}`))
				return
			}
			logins++
			_, _ = fmt.Fprintf(w, `{"session":{"valid":true,"sid":"sid-%d","validity":1800}}`, logins)
			return
		}
		// The first session expires after its first use.
		if r.Header.Get("X-FTL-SID") != fmt.Sprintf("sid-%d", logins) || (logins == 1 && r.URL.Path != "/api/stats/summary") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/stats/summary":
			_, _ = w.Write([]byte(`{"queries":{"total":2000,"blocked":500,"percent_blocked":25.0}}`))
		case "/api/dns/blocking":
			_, _ = w.Write([]byte(`{"blocking":"disabled","timer":30}`))
		case "/api/stats/top_domains":
			if r.URL.Query().Get("blocked") != "true" {
				t.Errorf("top domains must be the blocked ones: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"domains":[{"domain":"ads.example","count":300},{"domain":"t.example","count":200}]}`))
		}
	}))
	defer srv.Close()
	pi := NewPiHole(srv.Client())
	s := domain.Settings{Provider: domain.ProviderPiHole, URL: srv.URL, Password: "app-pw"}
	st, err := pi.Stats(context.Background(), s)
	if err != nil || st.Queries != 2000 || st.Blocked != 500 || st.ProtectionEnabled || st.HasAvgProcessingTime || st.PauseRemaining != 30*time.Second ||
		len(st.TopBlocked) != 2 || st.TopBlocked[0].Domain != "ads.example" {
		t.Fatalf("unexpected %+v %v", st, err)
	}
	if logins != 2 {
		t.Fatalf("an expired session is renewed once, got %d logins", logins)
	}
	if _, err := pi.Stats(context.Background(), s); err != nil || logins != 2 {
		t.Fatalf("the session is reused: %v, %d logins", err, logins)
	}
	s.Password = "wrong"
	if _, err := pi.Stats(context.Background(), s); err == nil || !strings.Contains(err.Error(), "password") {
		t.Fatalf("expected a password error, got %v", err)
	}
}

func TestPiHoleWithoutPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth" || r.Header.Get("X-FTL-SID") != "" {
			t.Errorf("no session without a password: %s", r.URL.Path)
		}
		switch r.URL.Path {
		case "/api/stats/summary":
			_, _ = w.Write([]byte(`{"queries":{"total":10,"blocked":1}}`))
		case "/api/dns/blocking":
			_, _ = w.Write([]byte(`{"blocking":"enabled"}`))
		default:
			_, _ = w.Write([]byte(`{"domains":[]}`))
		}
	}))
	defer srv.Close()
	st, err := NewPiHole(srv.Client()).Stats(context.Background(), domain.Settings{Provider: domain.ProviderPiHole, URL: srv.URL})
	if err != nil || !st.ProtectionEnabled || st.Queries != 10 {
		t.Fatalf("unexpected %+v %v", st, err)
	}
}

func TestPause(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = append(got, r.Method+" "+r.URL.Path+" "+string(body))
		switch r.URL.Path {
		case "/control/protection":
			if u, p, _ := r.BasicAuth(); u != "admin" || p != "pw" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte("OK"))
		case "/api/dns/blocking":
			_, _ = w.Write([]byte(`{"blocking":"disabled","timer":300}`))
		}
	}))
	defer srv.Close()
	ag := NewAPI(srv.Client())
	if err := ag.Pause(context.Background(), domain.Settings{URL: srv.URL, Username: "admin", Password: "pw"}, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := ag.Pause(context.Background(), domain.Settings{URL: srv.URL, Username: "admin", Password: "no"}, time.Minute); err == nil {
		t.Fatal("expected an auth error")
	}
	pi := NewPiHole(srv.Client())
	if err := pi.Pause(context.Background(), domain.Settings{URL: srv.URL}, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	if got[0] != `POST /control/protection {"duration":300000,"enabled":false}` || got[2] != `POST /api/dns/blocking {"blocking":false,"timer":300}` {
		t.Fatalf("%q", got)
	}
}
