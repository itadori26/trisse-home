package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"trisse-home/internal/shared/httpapi"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// RequestLog logs every API request with method, path, status and duration.
// Static asset requests are logged at debug level only to keep logs readable.
func RequestLog(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			level := slog.LevelInfo
			if !strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
				level = slog.LevelDebug
			}
			if rec.status >= http.StatusInternalServerError {
				level = slog.LevelError
			}
			log.Log(r.Context(), level, "http request",
				"method", r.Method, "path", r.URL.Path, "status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"client_ip", httpapi.ClientIP(r.Context()).String())
		})
	}
}
