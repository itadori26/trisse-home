package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// MaxJSONBodyBytes limits request bodies so a client cannot exhaust memory.
const MaxJSONBodyBytes = 1 << 20

// ErrBadRequest is returned by DecodeJSON for malformed payloads.
var ErrBadRequest = errors.New("malformed request body")

// WriteJSON encodes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode json response", "err", err)
	}
}

// DecodeJSON strictly decodes the request body into dst. Unknown fields are
// rejected so client bugs surface early instead of being silently ignored.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: %w", ErrBadRequest, err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: trailing data", ErrBadRequest)
	}
	return nil
}
