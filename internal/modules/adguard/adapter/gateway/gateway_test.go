package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s, _ := domain.NewSettings("http://ag:3000", "admin", "s3cret")
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT password_encrypted FROM adguard_settings`).Scan(&stored)
	if stored == "s3cret" || strings.Contains(stored, "s3cret") {
		t.Fatalf("password stored in plain text: %q", stored)
	}
	if got, err := repo.Load(ctx); err != nil || got != s {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

func TestAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, p, ok := r.BasicAuth(); !ok || u != "admin" || p != "pw" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/control/stats":
			_, _ = w.Write([]byte(`{"num_dns_queries":1000,"num_blocked_filtering":150,"avg_processing_time":0.0042,
				"top_blocked_domains":[{"ads.example":40},{"tracker.example":90}]}`))
		case "/control/status":
			_, _ = w.Write([]byte(`{"protection_enabled":true}`))
		}
	}))
	defer srv.Close()
	api := NewAPI(srv.Client())
	st, err := api.Stats(context.Background(), domain.Settings{URL: srv.URL, Username: "admin", Password: "pw"})
	if err != nil || st.Queries != 1000 || !st.ProtectionEnabled || st.TopBlocked[0].Domain != "tracker.example" || st.AvgProcessingTime.Microseconds() != 4200 {
		t.Fatalf("unexpected %+v %v", st, err)
	}
	if _, err := api.Stats(context.Background(), domain.Settings{URL: srv.URL, Username: "admin", Password: "bad"}); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 error, got %v", err)
	}
}
