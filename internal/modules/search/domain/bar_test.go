package domain

import (
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestDefaultBarOptions(t *testing.T) {
	o := DefaultBarOptions()
	if o.OpenInNewTab() || o.Prefixes() != (CapturePrefixes{Todo: "t", Note: "n", Bookmark: "b"}) {
		t.Fatalf("%+v", o)
	}
}

func TestNewBarOptions(t *testing.T) {
	o, err := NewBarOptions(true, CapturePrefixes{Todo: " A ", Note: "nz", Bookmark: "ö1"})
	if err != nil {
		t.Fatal(err)
	}
	if !o.OpenInNewTab() || o.Prefixes() != (CapturePrefixes{Todo: "a", Note: "nz", Bookmark: "ö1"}) {
		t.Fatalf("trimmed and lower-cased: %+v", o.Prefixes())
	}
	for name, p := range map[string]CapturePrefixes{
		"empty":       {Todo: "", Note: "n", Bookmark: "b"},
		"too long":    {Todo: "todos", Note: "n", Bookmark: "b"},
		"space":       {Todo: "t o", Note: "n", Bookmark: "b"},
		"punctuation": {Todo: "t:", Note: "n", Bookmark: "b"},
		"bang":        {Todo: "!w", Note: "n", Bookmark: "b"},
		"duplicate":   {Todo: "t", Note: "T", Bookmark: "b"},
		"duplicate 2": {Todo: "t", Note: "n", Bookmark: "n"},
	} {
		if _, err := NewBarOptions(false, p); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: want a validation error, got %v", name, err)
		}
	}
}
