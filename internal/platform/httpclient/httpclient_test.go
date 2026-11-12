package httpclient

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
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
