package middleware

import (
	"log/slog"
	"net/http"
	"net/netip"
	"sync"

	"trisse-home/internal/shared/httpapi"
)

// maxRememberedClients bounds the memory used to deduplicate allow logs.
const maxRememberedClients = 1024

// AllowList rejects requests whose client IP is outside the allowed prefixes.
// Denied requests are always logged; allowed clients are logged once when
// first seen so the allowlist can be audited without flooding the log.
func AllowList(allowed []netip.Prefix, log *slog.Logger) Middleware {
	var mu sync.Mutex
	seen := make(map[netip.Addr]struct{})
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := httpapi.ClientIP(r.Context())
			// Loopback is always allowed: it can only originate from the host or
			// container itself, e.g. the container health check.
			if !ip.IsLoopback() && !contains(allowed, ip) {
				log.WarnContext(r.Context(), "client ip denied by allowlist",
					"client_ip", ip.String(), "remote_addr", r.RemoteAddr, "path", r.URL.Path)
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			mu.Lock()
			_, known := seen[ip]
			if !known && len(seen) < maxRememberedClients {
				seen[ip] = struct{}{}
			}
			mu.Unlock()
			if !known {
				log.InfoContext(r.Context(), "client ip allowed (first request)", "client_ip", ip.String())
			}
			next.ServeHTTP(w, r)
		})
	}
}
