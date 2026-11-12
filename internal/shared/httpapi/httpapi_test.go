package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestWriteDomainError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{"validation", kernel.NewValidationError("name", "too long"), http.StatusUnprocessableEntity, "too long"},
		{"public", kernel.NewPublicError(kernel.ErrConflict, "exists"), http.StatusConflict, "exists"},
		{"wrapped not found", fmt.Errorf("x: %w", kernel.ErrNotFound), http.StatusNotFound, "Not Found"},
		{"bad request", ErrBadRequest, http.StatusBadRequest, "The request body is malformed."},
		{"internal hides details", errors.New("db password wrong"), http.StatusInternalServerError, "An internal error occurred."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			WriteDomainError(rec, httptest.NewRequest(http.MethodGet, "/", nil), tt.err)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			var body ErrorBody
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Message != tt.wantMsg {
				t.Fatalf("message = %q, want %q", body.Error.Message, tt.wantMsg)
			}
		})
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	var dst struct {
		A string `json:"a"`
	}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"x","b":1}`))
	if err := DecodeJSON(httptest.NewRecorder(), req, &dst); !errors.Is(err, ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":"x"}`))
	if err := DecodeJSON(httptest.NewRecorder(), req, &dst); err != nil || dst.A != "x" {
		t.Fatalf("unexpected result %v %q", err, dst.A)
	}
}

func TestWriteDomainErrorCarriesMessageKey(t *testing.T) {
	rec := httptest.NewRecorder()
	err := kernel.NewValidationErrorf("file", "At most {max} bookmarks can be imported at once.", "max", 500)
	WriteDomainError(rec, httptest.NewRequest(http.MethodPost, "/", nil), err)
	var body ErrorBody
	_ = json.NewDecoder(rec.Body).Decode(&body)
	d := body.Error
	if d.Message != "At most 500 bookmarks can be imported at once." || d.MessageKey != kernel.MessageKey("At most {max} bookmarks can be imported at once.") ||
		d.Params["max"] != "500" || d.Field != "file" {
		t.Fatalf("%+v", d)
	}
}
