package main

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"trisse-home/internal/platform/metrics"
	"trisse-home/internal/shared/kernel"
)

// metricsHandler serves GET /metrics in the Prometheus text format (#109).
// It is only registered with METRICS_ENABLED=true, requires the bearer
// token METRICS_TOKEN and is additionally limited by ALLOWED_CIDRS (global
// middleware). Modules contribute through kernel.MetricSource and never see
// Prometheus types.
func metricsHandler(a *app, httpStats *metrics.HTTPStats) http.Handler {
	want := []byte("Bearer " + a.cfg.MetricsToken)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := []byte(strings.TrimSpace(r.Header.Get("Authorization")))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="trisse-home metrics"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		samples := metrics.Runtime(version)
		samples = append(samples, httpStats.Samples()...)
		for _, m := range a.registry.All() {
			if src, ok := m.(kernel.MetricSource); ok {
				samples = append(samples, src.Metrics(r.Context())...)
			}
		}
		w.Header().Set("Content-Type", metrics.ContentType)
		w.Header().Set("Cache-Control", "no-store")
		_ = metrics.Write(w, samples) // the client went away
	})
}
