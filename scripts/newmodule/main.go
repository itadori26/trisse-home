// Command newmodule creates a new feature module from templates: backend
// (internal/modules/<name>, all four rings with tests and a migration) and
// frontend (web/src/features/<name> with a widget), and registers it in the
// composition roots, tygo.yaml and both message catalogues.
//
// Usage: go run ./scripts/newmodule -name recipes [-title "Recipes"]
// (or: make new-module NAME=recipes)
package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates
var templates embed.FS

// Data is passed to every template.
type Data struct {
	Name   string // package and URL name, e.g. "recipes"
	Pascal string // "Recipes"
	Title  string // human-readable, e.g. "Recipes"
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]{1,19}$`)

// file maps a template to its output path (relative to the repository root).
type file struct{ tmpl, out string }

func files(d Data) []file {
	b, f := "internal/modules/"+d.Name, "web/src/features/"+d.Name
	return []file{
		{"backend/doc.go.tmpl", b + "/doc.go"},
		{"backend/domain_doc.go.tmpl", b + "/domain/doc.go"},
		{"backend/domain.go.tmpl", b + "/domain/item.go"},
		{"backend/domain_test.go.tmpl", b + "/domain/item_test.go"},
		{"backend/usecase_doc.go.tmpl", b + "/usecase/doc.go"},
		{"backend/usecase.go.tmpl", b + "/usecase/usecase.go"},
		{"backend/usecase_test.go.tmpl", b + "/usecase/usecase_test.go"},
		{"backend/migration.sql.tmpl", b + "/adapter/gateway/migrations/0001_init.sql"},
		{"backend/gateway.go.tmpl", b + "/adapter/gateway/sqlite.go"},
		{"backend/gateway_test.go.tmpl", b + "/adapter/gateway/sqlite_test.go"},
		{"backend/presenter.go.tmpl", b + "/adapter/presenter/presenter.go"},
		{"backend/dto.go.tmpl", b + "/adapter/controller/dto.go"},
		{"backend/handlers.go.tmpl", b + "/adapter/controller/handlers.go"},
		{"backend/handlers_test.go.tmpl", b + "/adapter/controller/handlers_test.go"},
		{"backend/module.go.tmpl", b + "/module.go"},
		{"frontend/domain.ts.tmpl", f + "/domain/" + d.Name + ".ts"},
		{"frontend/domain.test.ts.tmpl", f + "/domain/" + d.Name + ".test.ts"},
		{"frontend/usecase.ts.tmpl", f + "/usecase/" + d.Name + "UseCases.ts"},
		{"frontend/usecase.test.ts.tmpl", f + "/usecase/" + d.Name + "UseCases.test.ts"},
		{"frontend/api.ts.tmpl", f + "/adapter/api/" + d.Name + "Api.ts"},
		{"frontend/widget.svelte.tmpl", f + "/ui/" + d.Pascal + "Widget.svelte"},
		{"frontend/index.ts.tmpl", f + "/index.ts"},
	}
}

// render executes one template; Go output is gofmt-ed.
func render(name string, d Data) ([]byte, error) {
	t, err := template.ParseFS(templates, "templates/"+name)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, d); err != nil {
		return nil, err
	}
	if strings.HasSuffix(name, ".go.tmpl") {
		return format.Source(buf.Bytes())
	}
	return buf.Bytes(), nil
}

// insertBefore inserts text on a new line before the line containing marker.
func insertBefore(src, marker, text string) (string, error) {
	i := strings.Index(src, marker)
	if i < 0 {
		return "", fmt.Errorf("marker %q not found", marker)
	}
	lineStart := strings.LastIndex(src[:i], "\n") + 1
	return src[:lineStart] + text + src[lineStart:], nil
}

// insertAfterLastImport adds an import line after the last top-level import.
func insertAfterLastImport(src, line string) string {
	lines := strings.Split(src, "\n")
	last := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "import ") {
			last = i
		}
	}
	lines = append(lines[:last+1], append([]string{line}, lines[last+1:]...)...)
	return strings.Join(lines, "\n")
}

type edit func(string, Data) (string, error)

var registrations = map[string]edit{
	"cmd/server/modules.go": func(s string, d Data) (string, error) {
		s, err := insertBefore(s, "// scaffold:modules", fmt.Sprintf("\t\t{%q, func() modules.Module { return %s.New(db, clk, ids) }},\n", d.Name, d.Name))
		if err != nil {
			return "", err
		}
		s = strings.Replace(s, "\t\"trisse-home/internal/modules\"\n", "\t\"trisse-home/internal/modules\"\n\t\"trisse-home/internal/modules/"+d.Name+"\"\n", 1)
		out, err := format.Source([]byte(s))
		return string(out), err
	},
	"tygo.yaml": func(s string, d Data) (string, error) {
		return strings.TrimRight(s, "\n") + fmt.Sprintf(`
  - path: "trisse-home/internal/modules/%[1]s/adapter/presenter"
    output_path: "web/src/shared/api-client/generated/%[1]s.ts"
  - path: "trisse-home/internal/modules/%[1]s/adapter/controller"
    output_path: "web/src/shared/api-client/generated/%[1]s-requests.ts"
    include_files: ["dto.go"]
`, d.Name), nil
	},
	"web/src/app/container.ts": func(s string, d Data) (string, error) {
		s = insertAfterLastImport(s, fmt.Sprintf("import { %[1]sApi, %[1]sUseCases } from '../features/%[2]s';", d.Pascal, d.Name))
		s, err := insertBefore(s, "// scaffold:container-fields", fmt.Sprintf("  %s: %sUseCases;\n", d.Name, d.Pascal))
		if err != nil {
			return "", err
		}
		return insertBefore(s, "// scaffold:container-init", fmt.Sprintf("    %s: new %sUseCases(new %sApi(http)),\n", d.Name, d.Pascal, d.Pascal))
	},
	"web/src/app/widgets.ts": func(s string, d Data) (string, error) {
		s = insertAfterLastImport(s, fmt.Sprintf("import { %sWidget } from '../features/%s';", d.Pascal, d.Name))
		return insertBefore(s, "// scaffold:widgets", fmt.Sprintf(`  {
    type: {
      kind: '%[1]s',
      get title() {
        return t('widget.%[1]s.title');
      },
      get description() {
        return t('widget.%[1]s.description');
      },
      defaultSize: { w: 4, h: 3 },
    },
    view: renderer(%[2]sWidget, (_config, c) => ({ %[1]s: c.%[1]s })),
  },
`, d.Name, d.Pascal))
	},
	"web/src/shared/i18n/locales/en.ts": func(s string, d Data) (string, error) {
		return insertBefore(s, "} as const;", messages(d, map[string]string{
			"widget." + d.Name + ".title":       d.Title,
			"widget." + d.Name + ".description": "Example widget of the " + d.Title + " module.",
			d.Name + ".add_placeholder":         "New item…",
			d.Name + ".remove":                  "Remove {title}",
			d.Name + ".empty":                   "No items yet.",
		}))
	},
	"web/src/shared/i18n/locales/de.ts": func(s string, d Data) (string, error) {
		i := strings.LastIndex(s, "};")
		if i < 0 {
			return "", errors.New("end of the German catalogue not found")
		}
		return s[:i] + messages(d, map[string]string{
			"widget." + d.Name + ".title":       d.Title,
			"widget." + d.Name + ".description": "Beispiel-Widget des Moduls " + d.Title + ".",
			d.Name + ".add_placeholder":         "Neuer Eintrag …",
			d.Name + ".remove":                  "{title} entfernen",
			d.Name + ".empty":                   "Noch keine Einträge.",
		}) + s[i:], nil
	},
}

func messages(d Data, m map[string]string) string {
	keys := []string{"widget." + d.Name + ".title", "widget." + d.Name + ".description", d.Name + ".add_placeholder", d.Name + ".remove", d.Name + ".empty"}
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "  %q: %q,\n", k, m[k])
	}
	return b.String()
}

func run(root string, d Data) error {
	if !namePattern.MatchString(d.Name) {
		return errors.New("the name must be 2-20 lowercase letters or digits and start with a letter")
	}
	if _, err := os.Stat(filepath.Join(root, "internal/modules", d.Name)); err == nil {
		return fmt.Errorf("module %q already exists", d.Name)
	}
	for _, f := range files(d) {
		out, err := render(f.tmpl, d)
		if err != nil {
			return fmt.Errorf("%s: %w", f.tmpl, err)
		}
		path := filepath.Join(root, f.out)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { //nolint:gosec // source directories are world-readable
			return err
		}
		if err := os.WriteFile(path, out, 0o644); err != nil { //nolint:gosec // source files are world-readable
			return err
		}
	}
	for path, change := range registrations {
		full := filepath.Join(root, path)
		src, err := os.ReadFile(full) //nolint:gosec // fixed paths inside the repository
		if err != nil {
			return err
		}
		out, err := change(string(src), d)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if err := os.WriteFile(full, []byte(out), 0o644); err != nil { //nolint:gosec // source files are world-readable
			return err
		}
	}
	return nil
}

func main() {
	name := flag.String("name", "", "module name, e.g. recipes")
	title := flag.String("title", "", "human-readable title (default: name with capital letter)")
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	d := Data{Name: *name, Pascal: strings.ToUpper((*name)[:min(1, len(*name))]) + (*name)[min(1, len(*name)):]}
	d.Title = *title
	if d.Title == "" {
		d.Title = d.Pascal
	}
	if err := run(*root, d); err != nil {
		fmt.Fprintln(os.Stderr, "newmodule:", err)
		os.Exit(1)
	}
	fmt.Printf("Created module %s. Next: make generate && (cd web && npx prettier --write src) && make ci\n", d.Name)
}
