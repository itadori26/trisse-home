package kernel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
)

// Errors of calls to owner-configured targets (#191). Gateways wrap them so
// TargetFailure can tell the user what went wrong without revealing raw
// error text (host, port, resolver details) to the browser.
var (
	// ErrTargetAuth: the target rejected the credentials (401/403).
	ErrTargetAuth = errors.New("the server rejected the credentials")
	// ErrTargetNotFound: the server answered, but not the expected path (404).
	ErrTargetNotFound = errors.New("the path was not found on the server")
	// ErrTargetBlocked: the address is not allowed for integrations.
	ErrTargetBlocked = errors.New("connections to this address are not allowed")
)

// TargetFailure turns an error of a call to an owner-configured target into
// a public error with a generic, translatable message. The original error
// is kept as Cause: it is logged, never shown.
func TargetFailure(kind error, cause error) *PublicError {
	var (
		netErr  net.Error
		dnsErr  *net.DNSError
		opErr   *net.OpError
		unknown x509.UnknownAuthorityError
		invalid x509.CertificateInvalidError
		host    x509.HostnameError
		verify  *tls.CertificateVerificationError
	)
	// Literal messages per cause, so they can be translated.
	var e *PublicError
	switch {
	case errors.Is(cause, ErrTargetBlocked):
		e = NewPublicError(kind, "Connections to this address are not allowed.")
	case errors.Is(cause, ErrTargetAuth):
		e = NewPublicError(kind, "The server rejected the login. Check user name, password or token.")
	case errors.Is(cause, ErrTargetNotFound):
		e = NewPublicError(kind, "The server answered, but the path was not found. Check the address.")
	case errors.As(cause, &unknown), errors.As(cause, &invalid), errors.As(cause, &host), errors.As(cause, &verify):
		e = NewPublicError(kind, "The certificate of the server is not trusted.")
	case errors.Is(cause, context.DeadlineExceeded), errors.As(cause, &netErr) && netErr.Timeout():
		e = NewPublicError(kind, "The server did not answer in time.")
	case errors.As(cause, &dnsErr), errors.As(cause, &opErr):
		e = NewPublicError(kind, "The server could not be reached. Check the address and port.")
	default:
		e = NewPublicError(kind, "The server sent an unexpected answer.")
	}
	e.Cause = cause
	return e
}
