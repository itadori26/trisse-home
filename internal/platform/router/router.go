// Package router implements modules.Router on top of the standard library
// ServeMux and adds the health check and the API not-found fallback.
//
// Ring: Frameworks & Drivers.
package router

import (
	"context"
	"net/http"

	"trisse-home/internal/shared/httpapi"
)

// Guard wraps handlers that require an authenticated session.
type Guard func(http.Handler) http.Handler

// Router dispatches requests to public and protected handlers.
type Router struct {
	mux   *http.ServeMux
	guard Guard
}

// New returns a Router that protects handlers with guard.
func New(guard Guard) *Router {
	r := &Router{mux: http.NewServeMux(), guard: guard}
	r.mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "Unknown API endpoint.")
	})
	return r
}

// Public registers a handler that is reachable without a session.
func (r *Router) Public(pattern string, h http.HandlerFunc) { r.mux.Handle(pattern, h) }

// Protected registers a handler behind the session guard.
func (r *Router) Protected(pattern string, h http.HandlerFunc) { r.mux.Handle(pattern, r.guard(h)) }

// Handle registers a plain handler, e.g. for static files.
func (r *Router) Handle(pattern string, h http.Handler) { r.mux.Handle(pattern, h) }

// ServeHTTP implements http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) { r.mux.ServeHTTP(w, req) }

// HealthCheck returns a handler that reports 200 when ping succeeds.
func HealthCheck(ping func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ping(r.Context()); err != nil {
			httpapi.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
