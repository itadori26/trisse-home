package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollect(t *testing.T) {
	dir := t.TempDir()
	src := `package x

import "trisse-home/internal/shared/kernel"

var A = kernel.NewPublicError(kernel.ErrInvalid, "Plain " +
	"message.")
var B = kernel.NewValidationErrorf("f", "At most {max}.", "max", 3)

func bad(s string) error { return kernel.NewPublicError(kernel.ErrInvalid, "Bad "+s) }
`
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	messages, problems, err := collect([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || len(problems) != 1 || !strings.Contains(problems[0], "x.go:9") {
		t.Fatalf("%v %v", messages, problems)
	}
	out := string(render(messages))
	if !strings.Contains(out, `"Plain message."`) || !strings.Contains(out, "ServerMessageKey") {
		t.Fatal(out)
	}
}
