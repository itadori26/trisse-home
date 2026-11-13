package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestCollect(t *testing.T) {
	dir := t.TempDir()
	src := `package x

import "trisse-home/internal/shared/kernel"

var A = kernel.NewPublicError(kernel.ErrInvalid, "Plain " +
	"message.")
var B = kernel.NewValidationErrorf("f", "At most {max}.", "max", 3)

func bad(s string) error { return kernel.NewPublicError(kernel.ErrInvalid, "Bad "+s) }
`
	writeFile(t, filepath.Join(dir, "x.go"), src)
	messages, _, problems, err := collect([]string{dir})
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

func TestMessagesAreSplitByOwner(t *testing.T) {
	dir := t.TempDir()
	const head = "package x\n\nimport \"trisse-home/internal/shared/kernel\"\n\n"
	writeFile(t, filepath.Join(dir, "internal/modules/media/a.go"),
		head+`var A = kernel.NewPublicError(kernel.ErrInvalid, "Only media.")
var C = kernel.NewPublicError(kernel.ErrInvalid, "Media and arr.")
`)
	writeFile(t, filepath.Join(dir, "internal/modules/arr/a.go"),
		head+`var C = kernel.NewPublicError(kernel.ErrInvalid, "Media and arr.")
`)
	writeFile(t, filepath.Join(dir, "internal/shared/kernel/e.go"),
		head+`var D = kernel.NewPublicError(kernel.ErrInvalid, "Kernel only.")
`)

	messages, users, problems, err := collect([]string{dir})
	if err != nil || len(problems) != 0 {
		t.Fatal(err, problems)
	}
	byOwner := split(messages, users)
	has := func(owner, text string) bool {
		for _, m := range byOwner[owner] {
			if m == text {
				return true
			}
		}
		return false
	}
	if !has("media", "Only media.") || !has("shared", "Media and arr.") || !has("shared", "Kernel only.") {
		t.Fatalf("unexpected split: %v", byOwner)
	}
	if len(byOwner["arr"]) != 0 {
		t.Fatalf("a message shared with another module belongs to shared: %v", byOwner)
	}

	out := filepath.Join(dir, "out")
	writeFile(t, filepath.Join(out, "gone.ts"), "stale")
	if err := write(out, byOwner); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"media.ts", "shared.ts"} {
		if _, err := os.Stat(filepath.Join(out, want)); err != nil {
			t.Errorf("%s was not written: %v", want, err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "gone.ts")); err == nil {
		t.Error("a file of an owner without messages must be removed")
	}
}
