package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every module that registers itself from cmd/server/module_<name>.go documents
// itself in its own files (#194): docs/modules/<name>.md and the OpenAPI
// fragment api/modules/<name>.yaml. That keeps the shared README and OpenAPI
// document free of per-module edits.
func TestSelfRegisteredModulesAreDocumented(t *testing.T) {
	files, err := filepath.Glob("module_*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no module files found: %v", err)
	}
	for _, f := range files {
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(f), "module_"), ".go")
		for _, doc := range []string{"docs/modules/" + name + ".md", "api/modules/" + name + ".yaml"} {
			if _, err := os.Stat(filepath.Join("..", "..", doc)); err != nil {
				t.Errorf("module %s: %s is missing (make new-module creates it)", name, doc)
			}
		}
	}
}
