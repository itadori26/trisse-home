package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestResolveClientIP(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("172.18.0.0/16")}
	tests := []struct {
		name, remote, xff, want string
	}{
		{"direct client ignores xff", "192.168.178.20:5000", "1.2.3.4", "192.168.178.20"},
		{"trusted proxy uses xff", "172.18.0.5:1234", "192.168.178.30", "192.168.178.30"},
		{"rightmost untrusted wins", "172.18.0.5:1234", "8.8.8.8, 10.13.13.2, 172.18.0.9", "10.13.13.2"},
		{"trusted proxy without xff", "172.18.0.5:1234", "", "172.18.0.5"},
		{"ipv4 mapped ipv6", "[::ffff:192.168.178.2]:80", "", "192.168.178.2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remote
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			if got := ResolveClientIP(r, trusted); got.String() != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAllowList(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	allowed := []netip.Prefix{netip.MustParsePrefix("192.168.178.0/24")}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }),
		ClientIP(nil), AllowList(allowed, log))
	for remote, want := range map[string]int{
		"192.168.178.7:1": http.StatusNoContent,
		"10.0.0.1:1":      http.StatusForbidden,
		"garbage":         http.StatusForbidden,
		"127.0.0.1:1":     http.StatusNoContent,
	} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = remote
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		if rec.Code != want {
			t.Errorf("%s: status %d, want %d", remote, rec.Code, want)
		}
	}
}

func TestRecover(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := Recover(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}
