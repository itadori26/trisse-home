package kernel

import (
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
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
	// Template is the English message with {name} placeholders and Params
	// their values; clients translate MessageKey(Template) (see #102).
	Template string
	Params   map[string]string
}

// NewValidationError returns a ValidationError for field with a user-facing
// message. The message must be a string literal so it can be translated.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message, Template: message}
}

// NewValidationErrorf is NewValidationError with placeholders: template
// contains {name} and params are name/value pairs.
func NewValidationErrorf(field, template string, params ...any) *ValidationError {
	p := pairs(params)
	return &ValidationError{Field: field, Message: fill(template, p), Template: template, Params: p}
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
	// Template and Params: see ValidationError.
	Template string
	Params   map[string]string
}

// NewPublicError returns an error of the given kind with a user-facing
// message. The message must be a string literal so it can be translated.
func NewPublicError(kind error, message string) *PublicError {
	return &PublicError{Kind: kind, Message: message, Template: message}
}

// NewPublicErrorf is NewPublicError with placeholders: template contains
// {name} and params are name/value pairs.
func NewPublicErrorf(kind error, template string, params ...any) *PublicError {
	p := pairs(params)
	return &PublicError{Kind: kind, Message: fill(template, p), Template: template, Params: p}
}

// Error implements the error interface.
func (e *PublicError) Error() string { return e.Message }

// Unwrap exposes the error kind to errors.Is.
func (e *PublicError) Unwrap() error { return e.Kind }

// MessageKey identifies a message template independently of its language:
// "srv_" and the 64-bit FNV-1a hash of the template in hex. It is an
// identifier, not a security feature. The frontend catalogue is generated
// from the same templates (scripts/servermessages).
func MessageKey(template string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(template)) // hash.Hash never fails
	return fmt.Sprintf("srv_%016x", h.Sum64())
}

func pairs(params []any) map[string]string {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]string, len(params)/2)
	for i := 0; i+1 < len(params); i += 2 {
		out[fmt.Sprint(params[i])] = fmt.Sprint(params[i+1])
	}
	return out
}

func fill(template string, params map[string]string) string {
	for k, v := range params {
		template = strings.ReplaceAll(template, "{"+k+"}", v)
	}
	return template
}
