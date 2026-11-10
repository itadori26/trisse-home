package kernel

import (
	"errors"
	"testing"
)

func TestValidationErrorIsInvalid(t *testing.T) {
	err := NewValidationError("title", "must not be empty")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ValidationError to match ErrInvalid")
	}
	if got, want := err.Error(), "title: must not be empty"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestPublicErrorKeepsKind(t *testing.T) {
	err := NewPublicError(ErrConflict, "already set up")
	if !errors.Is(err, ErrConflict) || errors.Is(err, ErrNotFound) {
		t.Fatalf("unexpected kind matching for %v", err)
	}
}
