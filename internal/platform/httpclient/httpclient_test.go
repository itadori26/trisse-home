package httpclient

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

func TestPinnedCertificate(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()
	sum := sha256.Sum256(srv.Certificate().Raw)
	fp := strings.ToUpper(hex.EncodeToString(sum[:]))
	withColons := make([]string, 0, len(fp)/2)
	for i := 0; i < len(fp); i += 2 {
		withColons = append(withColons, fp[i:i+2])
	}

	for name, pin := range map[string]string{"hex": fp, "colons": strings.Join(withColons, ":")} {
		res, err := New(WithPinnedCertificate(pin)).Get(srv.URL)
		if err != nil {
			t.Fatalf("%s: the pinned self-signed certificate must be accepted: %v", name, err)
		}
		_ = res.Body.Close()
	}
	for name, pin := range map[string]string{"other": strings.Repeat("ab", 32), "invalid": "nope"} {
		if res, err := New(WithPinnedCertificate(pin)).Get(srv.URL); err == nil {
			_ = res.Body.Close()
			t.Fatalf("%s: a different certificate must be rejected", name)
		}
	}
	if res, err := New().Get(srv.URL); err == nil {
		_ = res.Body.Close()
		t.Fatal("without pinning the self-signed certificate must be rejected")
	}
}

func TestBlocked(t *testing.T) {
	for addr, want := range map[string]bool{
		"127.0.0.1":        true,
		"::1":              true,
		"169.254.169.254":  true,
		"fe80::1":          true,
		"0.0.0.0":          true,
		"0.1.2.3":          true,
		"224.0.0.1":        true,
		"::ffff:127.0.0.1": true,
		"192.168.178.1":    false,
		"10.13.13.2":       false,
		"100.64.0.1":       false,
		"fd00::1":          false,
		"93.184.216.34":    false,
	} {
		if got := Blocked(netip.MustParseAddr(addr)); got != want {
			t.Errorf("%s: got %v, want %v", addr, got, want)
		}
	}
}

func TestGuardRefusesLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "secret") }))
	defer srv.Close()
	if res, err := New().Get(srv.URL); err != nil {
		t.Fatalf("plain client: %v", err)
	} else {
		_ = res.Body.Close()
	}
	_, err := New(WithInternalAddressGuard()).Get(srv.URL)
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("guarded client reached loopback: %v", err)
	}
}
