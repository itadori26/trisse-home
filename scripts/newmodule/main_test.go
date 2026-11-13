package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplatesRender(t *testing.T) {
	d := withKeys(Data{Name: "sample", Pascal: "Sample", Title: "Sample"})
	for _, f := range files(d) {
		out, err := render(f.tmpl, d)
		if err != nil {
			t.Fatalf("%s: %v", f.tmpl, err)
		}
		if strings.HasSuffix(f.out, ".go") {
			if _, err := parser.ParseFile(token.NewFileSet(), f.out, out, parser.ParseComments); err != nil {
				t.Errorf("%s does not parse: %v", f.out, err)
			}
		}
		if strings.Contains(string(out), "{{.") {
			t.Errorf("%s contains unrendered template code", f.out)
		}
	}
}

// A new module consists of files of its own only (#193): scaffolding it must
// not change or even read any existing file, and two modules must not share a
// single output path.
func TestRunCreatesOnlyNewFiles(t *testing.T) {
	root := t.TempDir()
	existing := filepath.Join(root, "cmd/server/modules.go")
	if err := os.MkdirAll(filepath.Dir(existing), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("untouched"), 0o600); err != nil {
		t.Fatal(err)
	}

	paths := map[string]string{}
	for _, name := range []string{"sample", "other"} {
		d := Data{Name: name, Pascal: strings.ToUpper(name[:1]) + name[1:], Title: name}
		if err := run(root, d); err != nil {
			t.Fatal(err)
		}
		for _, f := range files(withKeys(d)) {
			if prev, dup := paths[f.out]; dup {
				t.Errorf("%s is written for %s and %s", f.out, prev, name)
			}
			paths[f.out] = name
			if _, err := os.Stat(filepath.Join(root, f.out)); err != nil {
				t.Errorf("%s was not created: %v", f.out, err)
			}
		}
	}
	if got, _ := os.ReadFile(existing); string(got) != "untouched" { //nolint:gosec // test fixture
		t.Fatalf("an existing file was modified: %q", got)
	}

	register, err := os.ReadFile(filepath.Join(root, "cmd/server/module_sample.go")) //nolint:gosec // test fixture
	if err != nil || !strings.Contains(string(register), `registerModule("sample"`) {
		t.Fatalf("module_sample.go: %v\n%s", err, register)
	}
	// The German backend messages are keyed by the hash of the English text.
	de, err := os.ReadFile(filepath.Join(root, "web/src/shared/i18n/locales/server-de/sample.ts")) //nolint:gosec // test fixture
	if err != nil || !strings.Contains(string(de), withKeys(Data{Title: "sample"}).KeyNotFound+":") {
		t.Fatalf("server-de/sample.ts: %v\n%s", err, de)
	}
}

func TestRunRefusesToOverwrite(t *testing.T) {
	root := t.TempDir()
	d := Data{Name: "sample", Pascal: "Sample", Title: "Sample"}
	if err := run(root, d); err != nil {
		t.Fatal(err)
	}
	if err := run(root, d); err == nil {
		t.Fatal("scaffolding an existing module must fail")
	}
}

func TestRunRejectsBadNames(t *testing.T) {
	for _, n := range []string{"", "A", "my-module", "x", "toolongnamewithmanycharacters"} {
		if err := run(t.TempDir(), Data{Name: n}); err == nil {
			t.Errorf("%q accepted", n)
		}
	}
}
