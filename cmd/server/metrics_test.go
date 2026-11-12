package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"trisse-home/internal/platform/config"
)

const testMetricsToken = "0123456789abcdef0123456789abcdef"

func getMetrics(t *testing.T, url, token string) (int, string) {
	t.Helper()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url+"/metrics", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

func TestMetricsEndpoint(t *testing.T) {
	srv := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) {
		c.MetricsEnabled = true
		c.MetricsToken = testMetricsToken
	})
	if code, _ := getMetrics(t, srv.URL, ""); code != http.StatusUnauthorized {
		t.Fatalf("no token: %d", code)
	}
	if code, _ := getMetrics(t, srv.URL, "wrong"); code != http.StatusUnauthorized {
		t.Fatalf("wrong token: %d", code)
	}
	code, body := getMetrics(t, srv.URL, testMetricsToken)
	if code != http.StatusOK {
		t.Fatalf("%d %s", code, body)
	}
	for _, want := range []string{"# TYPE go_goroutines gauge", "trisse_build_info{", "trisse_http_requests_total{method=\"GET\",code=\"4xx\"}", "trisse_host_cpu_ratio"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in\n%s", want, body)
		}
	}
}

func TestMetricsDisabledByDefault(t *testing.T) {
	srv := newTestServer(t, "http://127.0.0.1:1")
	if code, body := getMetrics(t, srv.URL, testMetricsToken); code == http.StatusOK && strings.Contains(body, "go_goroutines") {
		t.Fatal("metrics must be off by default")
	}
}
