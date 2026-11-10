package httpapi

import (
	"context"
	"net/netip"
)

type contextKey int

const clientIPKey contextKey = iota

// WithClientIP stores the resolved client address in ctx. The platform
// middleware resolves it once so controllers never parse proxy headers.
func WithClientIP(ctx context.Context, ip netip.Addr) context.Context {
	return context.WithValue(ctx, clientIPKey, ip)
}

// ClientIP returns the client address stored by WithClientIP, or the zero
// Addr if none was stored.
func ClientIP(ctx context.Context) netip.Addr {
	ip, _ := ctx.Value(clientIPKey).(netip.Addr)
	return ip
}
