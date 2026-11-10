package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"

	"trisse-home/internal/shared/httpapi"
)

// ClientIP resolves the real client address once per request and stores it in
// the request context. X-Forwarded-For is only honoured when the direct peer
// is a trusted proxy; otherwise any client could spoof its address and bypass
// the allowlist.
func ClientIP(trustedProxies []netip.Prefix) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ResolveClientIP(r, trustedProxies)
			next.ServeHTTP(w, r.WithContext(httpapi.WithClientIP(r.Context(), ip)))
		})
	}
}

// ResolveClientIP returns the client address for r. When the peer is trusted,
// the X-Forwarded-For chain is walked from the right and the first untrusted
// address is returned.
func ResolveClientIP(r *http.Request, trustedProxies []netip.Prefix) netip.Addr {
	peer := parseAddr(r.RemoteAddr)
	if !contains(trustedProxies, peer) {
		return peer
	}
	hops := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	for i := len(hops) - 1; i >= 0; i-- {
		addr, err := netip.ParseAddr(strings.TrimSpace(hops[i]))
		if err != nil {
			break
		}
		addr = addr.Unmap()
		if !contains(trustedProxies, addr) {
			return addr
		}
	}
	return peer
}

func parseAddr(remote string) netip.Addr {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	return addr.Unmap()
}

func contains(prefixes []netip.Prefix, addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	for _, p := range prefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}
