// Package middleware contains the cross-cutting HTTP decorators of the
// platform: panic recovery, request logging, client IP resolution, the IP
// allowlist and security headers. Session and CSRF checks are provided by the
// auth module and plugged into the router by the composition root.
//
// Ring: Frameworks & Drivers.
package middleware

import "net/http"

// Middleware decorates an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain wraps h so that the first middleware is the outermost one.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
