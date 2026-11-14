package httpclient

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntegrationControl(t *testing.T) {
	for addr, blocked := range map[string]bool{
		"169.254.169.254:80":      true, // cloud metadata
		"[fe80::1]:443":           true,
		"[::ffff:169.254.1.1]:80": true,
		"0.0.0.0:80":              true,
		"[ff02::1]:80":            true,
		"127.0.0.1:8080":          false, // home-lab services on the same host
		"192.168.178.20:443":      false,
		"[fd00::5]:8006":          false,
	} {
		err := integrationControl("tcp", addr, nil)
		if blocked != errors.Is(err, ErrBlockedAddress) {
			t.Errorf("%s: blocked=%v, err=%v", addr, blocked, err)
		}
	}
}

func TestIntegrationGuardRefusesLinkLocal(t *testing.T) {
	// The dialer refuses before any packet is sent.
	_, err := New(WithIntegrationGuard()).Get("http://169.254.169.254/latest/meta-data/")
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("metadata address reached: %v", err)
	}
}

func TestIntegrationGuardDoesNotFollowRedirects(t *testing.T) {
	var followed bool
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/elsewhere", http.StatusFound)
	})
	mux.HandleFunc("/elsewhere", func(w http.ResponseWriter, _ *http.Request) { followed = true })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	res, err := New(WithIntegrationGuard()).Get(srv.URL) // loopback is allowed
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusFound || followed {
		t.Fatalf("status %d, followed %v", res.StatusCode, followed)
	}
}

func TestMaxResponseBytes(t *testing.T) {
	body := strings.Repeat("x", 100)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chunked" {
			w.(http.Flusher).Flush() // no Content-Length
		}
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	read := func(c *http.Client, path string) (string, error) {
		res, err := c.Get(srv.URL + path)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		return string(b), err
	}
	if got, err := read(New(WithIntegrationGuard(), WithMaxResponseBytes(100)), "/"); err != nil || got != body {
		t.Fatalf("exactly at the limit: %q %v", got, err)
	}
	if _, err := read(New(WithIntegrationGuard(), WithMaxResponseBytes(99)), "/"); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("content length above the limit: %v", err)
	}
	if _, err := read(New(WithIntegrationGuard(), WithMaxResponseBytes(50)), "/chunked"); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("streamed body above the limit: %v", err)
	}
}

func TestPrivateOnlyControl(t *testing.T) {
	for addr, blocked := range map[string]bool{
		"169.254.169.254:80":  true, // still refused like any integration
		"93.184.215.14:443":   true, // public
		"[2606:4700::1]:443":  true,
		"8.8.8.8:53":          true,
		"127.0.0.1:8080":      false,
		"[::1]:8080":          false,
		"10.0.0.5:80":         false,
		"172.16.3.4:80":       false,
		"192.168.178.20:443":  false,
		"100.100.1.2:80":      false, // CGNAT, e.g. Tailscale
		"[fd00::5]:8006":      false,
		"[::ffff:10.1.1.1]:1": false,
	} {
		err := privateControl("tcp", addr, nil)
		if blocked != errors.Is(err, ErrBlockedAddress) {
			t.Errorf("%s: blocked=%v, err=%v", addr, blocked, err)
		}
	}
}

func TestPrivateOnlyRefusesPublicBeforeConnecting(t *testing.T) {
	_, err := New(WithPrivateOnly()).Get("http://93.184.215.14/")
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("public address reached: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") }))
	defer srv.Close()
	res, err := New(WithPrivateOnly()).Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
}
