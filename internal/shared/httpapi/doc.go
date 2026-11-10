// Package httpapi contains small helpers shared by all HTTP controllers:
// JSON encoding, request decoding, the error envelope, mapping of domain error
// kinds to status codes, and request-scoped values such as the client IP.
//
// Ring: Interface Adapters. It may depend on the shared kernel and the
// standard library only.
package httpapi
