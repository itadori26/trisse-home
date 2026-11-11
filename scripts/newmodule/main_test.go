package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestTemplatesRender(t *testing.T) {
	d := Data{Name: "sample", Pascal: "Sample", Title: "Sample"}
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

func TestRegistrations(t *testing.T) {
	d := Data{Name: "sample", Pascal: "Sample", Title: "Sample"}
	modules := "package main\n\nimport (\n\t\"trisse-home/internal/modules\"\n)\n\nvar _ = []struct{}{\n\t\t// scaffold:modules\n}\n"
	out, err := registrations["cmd/server/modules.go"](modules, d)
	if err != nil || !strings.Contains(out, `"trisse-home/internal/modules/sample"`) || !strings.Contains(out, `sample.New(db, clk, ids)`) {
		t.Fatalf("modules.go: %v\n%s", err, out)
	}
	en, err := registrations["web/src/shared/i18n/locales/en.ts"]("export const en = {\n  \"a\": \"b\",\n} as const;\n", d)
	if err != nil || !strings.Contains(en, `"sample.remove": "Remove {title}"`) {
		t.Fatalf("en.ts: %v\n%s", err, en)
	}
	if _, err := registrations["web/src/app/widgets.ts"]("import x from 'y';\n];\n", d); err == nil {
		t.Fatal("missing marker must be an error")
	}
}

func TestRunRejectsBadNames(t *testing.T) {
	for _, n := range []string{"", "A", "my-module", "x", "toolongnamewithmanycharacters"} {
		if err := run(t.TempDir(), Data{Name: n}); err == nil {
			t.Errorf("%q accepted", n)
		}
	}
}
