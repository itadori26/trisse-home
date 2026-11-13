package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"trisse-home/internal/shared/kernel"
)

// ErrorBody is the JSON envelope for every error response.
type ErrorBody struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail carries a machine-readable code, a user-facing message and,
// for validation errors, the offending field.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	// MessageKey identifies the English template of Message so clients can
	// show a translation; Params fill its {placeholders}.
	MessageKey string            `json:"messageKey,omitempty"`
	Params     map[string]string `json:"params,omitempty"`
}

// WriteError writes an error envelope with the given status, code and message.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorBody{Error: ErrorDetail{Code: code, Message: message}})
}

type errorMapping struct {
	kind   error
	status int
	code   string
}

// mappings is ordered: the first matching kind wins.
var mappings = []errorMapping{
	{kernel.ErrInvalid, http.StatusUnprocessableEntity, "invalid"},
	{kernel.ErrNotFound, http.StatusNotFound, "not_found"},
	{kernel.ErrConflict, http.StatusConflict, "conflict"},
	{kernel.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
	{kernel.ErrForbidden, http.StatusForbidden, "forbidden"},
	{kernel.ErrTooManyRequests, http.StatusTooManyRequests, "too_many_requests"},
	{kernel.ErrUnavailable, http.StatusBadGateway, "unavailable"},
}

// WriteDomainError maps an error returned by a use case to an HTTP response.
// Only messages of errors explicitly marked as public reach the client;
// everything else is logged and answered with a generic message.
func WriteDomainError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrBadRequest) {
		WriteError(w, http.StatusBadRequest, "bad_request", "The request body is malformed.")
		return
	}
	for _, m := range mappings {
		if !errors.Is(err, m.kind) {
			continue
		}
		detail := ErrorDetail{Code: m.code, Message: http.StatusText(m.status)}
		var verr *kernel.ValidationError
		var perr *kernel.PublicError
		switch {
		case errors.As(err, &verr):
			detail.Message, detail.Field = verr.Message, verr.Field
			detail.MessageKey, detail.Params = messageKey(verr.Template), verr.Params
		case errors.As(err, &perr):
			detail.Message = perr.Message
			detail.MessageKey, detail.Params = messageKey(perr.Template), perr.Params
			if perr.Cause != nil {
				// Details of outbound calls stay in the log (#191).
				slog.WarnContext(r.Context(), "target call failed", "path", r.URL.Path, "err", perr.Cause)
			}
		}
		if m.status >= http.StatusInternalServerError {
			slog.WarnContext(r.Context(), "upstream failure", "path", r.URL.Path, "err", err)
		}
		WriteJSON(w, m.status, ErrorBody{Error: detail})
		return
	}
	slog.ErrorContext(r.Context(), "unhandled error", "path", r.URL.Path, "err", err)
	WriteError(w, http.StatusInternalServerError, "internal", "An internal error occurred.")
}

func messageKey(template string) string {
	if template == "" {
		return ""
	}
	return kernel.MessageKey(template)
}
