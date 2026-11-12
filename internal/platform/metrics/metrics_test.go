package metrics

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestWrite(t *testing.T) {
	var b strings.Builder
	err := Write(&b, []kernel.MetricSample{
		kernel.Gauge("trisse_service_up", "Whether the service answered.", 1, kernel.Label{Name: "service", Value: `NAS "main"`}),
		{Name: "trisse_a_total", Help: "Line\nbreak", Type: kernel.MetricCounter, Value: 3},
		kernel.Gauge("trisse_service_up", "ignored", 0, kernel.Label{Name: "service", Value: "Router"}),
		kernel.Gauge("bad name", "dropped", 1),
		kernel.Gauge("trisse_inf", "inf", math.Inf(1)),
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `# HELP trisse_a_total Line\nbreak
# TYPE trisse_a_total counter
trisse_a_total 3
# HELP trisse_inf inf
# TYPE trisse_inf gauge
trisse_inf +Inf
# HELP trisse_service_up Whether the service answered.
# TYPE trisse_service_up gauge
trisse_service_up{service="NAS \"main\""} 1
trisse_service_up{service="Router"} 0
`
	if b.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", b.String(), want)
	}
}

func TestRuntime(t *testing.T) {
	var b strings.Builder
	_ = Write(&b, Runtime("v1.2.3"))
	for _, s := range []string{"go_goroutines ", "go_gc_cycles_total ", "process_start_time_seconds ", `trisse_build_info{version="v1.2.3"`} {
		if !strings.Contains(b.String(), s) {
			t.Fatalf("missing %q in\n%s", s, b.String())
		}
	}
}

func TestHTTPStats(t *testing.T) {
	stats := NewHTTPStats()
	h := stats.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	for _, p := range []string{"/", "/", "/missing"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, p, nil))
	}
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("PROPFIND", "/", nil))
	var b strings.Builder
	_ = Write(&b, stats.Samples())
	out := b.String()
	for _, s := range []string{`trisse_http_requests_total{method="GET",code="2xx"} 2`, `{method="GET",code="4xx"} 1`, `{method="other",code="2xx"} 1`, "trisse_http_request_duration_seconds_count 4"} {
		if !strings.Contains(out, s) {
			t.Fatalf("missing %q in\n%s", s, out)
		}
	}
}
