package kernel

import (
	"errors"
	"fmt"
)

// Error kinds classify domain errors independently of any transport. Modules
// wrap these kinds into their own errors; controllers map kinds to status codes.
var (
	// ErrNotFound reports that a requested entity does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalid reports that input violates a business rule.
	ErrInvalid = errors.New("invalid")
	// ErrConflict reports that the operation conflicts with the current state.
	ErrConflict = errors.New("conflict")
	// ErrUnauthorized reports missing or wrong credentials.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden reports an operation that is not allowed in the current state.
	ErrForbidden = errors.New("forbidden")
	// ErrTooManyRequests reports that a rate limit was exceeded.
	ErrTooManyRequests = errors.New("too many requests")
	// ErrUnavailable reports that an external dependency could not be reached.
	ErrUnavailable = errors.New("unavailable")
)

// ValidationError describes a violated invariant on a specific field. Its
// message is safe to show to end users.
type ValidationError struct {
	Field   string
	Message string
}

// NewValidationError returns a ValidationError for field with a user-facing message.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Unwrap lets errors.Is(err, ErrInvalid) succeed for every ValidationError.
func (e *ValidationError) Unwrap() error { return ErrInvalid }

// PublicError is an error whose message may be shown to end users. Kind
// classifies it so transports can choose a status code.
type PublicError struct {
	Kind    error
	Message string
}

// NewPublicError returns an error of the given kind with a user-facing message.
func NewPublicError(kind error, message string) *PublicError {
	return &PublicError{Kind: kind, Message: message}
}

// Error implements the error interface.
func (e *PublicError) Error() string { return e.Message }

// Unwrap exposes the error kind to errors.Is.
func (e *PublicError) Unwrap() error { return e.Kind }
