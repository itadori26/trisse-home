// Command newmodule creates a new feature module from templates: backend
// (internal/modules/<name>, all four rings with tests and a migration),
// frontend (web/src/features/<name> with a widget) and the glue that makes
// the module known to the application.
//
// The glue lives in files of the module itself, so creating a module edits no
// existing file (#193):
//
//	cmd/server/module_<name>.go                       registers the backend module
//	cmd/server/api_<name>_test.go                     its end-to-end check (#194)
//	api/modules/<name>.yaml                           its OpenAPI fragment (#194)
//	web/src/app/widgets/<name>.ts                     widget definition
//	web/src/app/services/<name>.ts                    use cases in the container
//	web/src/features/<name>/i18n/{en,de}.ts           UI texts
//	web/src/shared/i18n/locales/server-de/<name>.ts   German backend messages
//	docs/modules/<name>.md                            documentation
//
// tygo and the backend message catalogue find the module by convention.
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

	"trisse-home/internal/shared/kernel"
)

//go:embed templates
var templates embed.FS

// Data is passed to every template.
type Data struct {
	Name   string // package and URL name, e.g. "recipes"
	Pascal string // "Recipes"
	Title  string // human-readable, e.g. "Recipes"
	// KeyNotFound and KeyTitle are the kernel.MessageKey of the two backend
	// messages of the domain template, for the German translations.
	KeyNotFound, KeyTitle string
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]{1,19}$`)

// withKeys fills the message keys of the domain template for d.Title.
func withKeys(d Data) Data {
	d.KeyNotFound = kernel.MessageKey("The " + d.Title + " item does not exist.")
	d.KeyTitle = kernel.MessageKey("The title of a " + d.Title + " item must be between 1 and 200 characters long.")
	return d
}

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
		{"backend/register.go.tmpl", "cmd/server/module_" + d.Name + ".go"},
		{"frontend/domain.ts.tmpl", f + "/domain/" + d.Name + ".ts"},
		{"frontend/domain.test.ts.tmpl", f + "/domain/" + d.Name + ".test.ts"},
		{"frontend/usecase.ts.tmpl", f + "/usecase/" + d.Name + "UseCases.ts"},
		{"frontend/usecase.test.ts.tmpl", f + "/usecase/" + d.Name + "UseCases.test.ts"},
		{"frontend/api.ts.tmpl", f + "/adapter/api/" + d.Name + "Api.ts"},
		{"frontend/widget.svelte.tmpl", f + "/ui/" + d.Pascal + "Widget.svelte"},
		{"frontend/index.ts.tmpl", f + "/index.ts"},
		{"frontend/i18n_en.ts.tmpl", f + "/i18n/en.ts"},
		{"frontend/i18n_de.ts.tmpl", f + "/i18n/de.ts"},
		{"frontend/app_widget.ts.tmpl", "web/src/app/widgets/" + d.Name + ".ts"},
		{"frontend/app_service.ts.tmpl", "web/src/app/services/" + d.Name + ".ts"},
		{"frontend/server_de.ts.tmpl", "web/src/shared/i18n/locales/server-de/" + d.Name + ".ts"},
		{"docs/module.md.tmpl", "docs/modules/" + d.Name + ".md"},
		{"api/module.yaml.tmpl", "api/modules/" + d.Name + ".yaml"},
		{"backend/e2e_test.go.tmpl", "cmd/server/api_" + d.Name + "_test.go"},
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

func run(root string, d Data) error {
	if !namePattern.MatchString(d.Name) {
		return errors.New("the name must be 2-20 lowercase letters or digits and start with a letter")
	}
	if _, err := os.Stat(filepath.Join(root, "internal/modules", d.Name)); err == nil {
		return fmt.Errorf("module %q already exists", d.Name)
	}
	d = withKeys(d)
	// Render everything first, so a template error leaves no half-created module.
	type output struct {
		path string
		data []byte
	}
	var outs []output
	for _, f := range files(d) {
		data, err := render(f.tmpl, d)
		if err != nil {
			return fmt.Errorf("%s: %w", f.tmpl, err)
		}
		outs = append(outs, output{filepath.Join(root, f.out), data})
	}
	for _, o := range outs {
		if _, err := os.Stat(o.path); err == nil {
			return fmt.Errorf("%s already exists", o.path)
		}
	}
	for _, o := range outs {
		if err := os.MkdirAll(filepath.Dir(o.path), 0o755); err != nil { //nolint:gosec // source directories are world-readable
			return err
		}
		if err := os.WriteFile(o.path, o.data, 0o644); err != nil { //nolint:gosec // source files are world-readable
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
	fmt.Printf("Created module %s; no existing file was changed.\n"+
		"Next: make generate && (cd web && npx prettier --write src) && make ci\n"+
		"Then adjust the picker category and icon in web/src/app/widgets/%s.ts and describe the module in docs/modules/%s.md.\n",
		d.Name, d.Name, d.Name)
}
