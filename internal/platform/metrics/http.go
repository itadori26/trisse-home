package metrics

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"trisse-home/internal/shared/kernel"
)

// HTTPStats counts HTTP requests by method and status class; labels are
// bounded (no paths) to keep the cardinality low.
type HTTPStats struct {
	mu       sync.Mutex
	requests map[[2]string]uint64 // method, code class
	seconds  float64
	count    uint64
}

// NewHTTPStats returns empty statistics.
func NewHTTPStats() *HTTPStats { return &HTTPStats{requests: map[[2]string]uint64{}} }

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

var knownMethods = map[string]bool{http.MethodGet: true, http.MethodPost: true, http.MethodPut: true, http.MethodPatch: true, http.MethodDelete: true, http.MethodHead: true}

// Middleware records every request.
func (s *HTTPStats) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		method := r.Method
		if !knownMethods[method] {
			method = "other"
		}
		class := strconv.Itoa(rec.status/100) + "xx"
		elapsed := time.Since(start).Seconds()
		s.mu.Lock()
		s.requests[[2]string{method, class}]++
		s.seconds += elapsed
		s.count++
		s.mu.Unlock()
	})
}

// Samples returns the HTTP server metrics.
func (s *HTTPStats) Samples() []kernel.MetricSample {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]kernel.MetricSample, 0, len(s.requests)+2)
	for k, n := range s.requests {
		out = append(out, kernel.MetricSample{Name: "trisse_http_requests_total", Help: "HTTP requests by method and status class.",
			Type: kernel.MetricCounter, Labels: []kernel.Label{{Name: "method", Value: k[0]}, {Name: "code", Value: k[1]}}, Value: float64(n)})
	}
	out = append(out,
		kernel.MetricSample{Name: "trisse_http_request_duration_seconds_sum", Help: "Total time spent serving HTTP requests.", Type: kernel.MetricCounter, Value: s.seconds},
		kernel.MetricSample{Name: "trisse_http_request_duration_seconds_count", Help: "Number of timed HTTP requests.", Type: kernel.MetricCounter, Value: float64(s.count)},
	)
	return out
}
