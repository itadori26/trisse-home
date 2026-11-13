package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const base = `openapi: 3.1.0
info:
  title: t
tags:
  - name: platform
paths:
  /healthz:
    get:
      tags: [platform]
      responses:
        "200": { $ref: "#/components/responses/Ok" }
components:
  responses:
    Ok: { description: OK }
  schemas:
    ErrorBody:
      type: object
`

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, frags map[string]string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "base.yaml", base)
	mods := filepath.Join(dir, "modules")
	_ = os.Mkdir(mods, 0o750)
	for name, content := range frags {
		write(t, mods, name+".yaml", content)
	}
	return build(filepath.Join(dir, "base.yaml"), mods)
}

const recipes = `# fragment of the recipes module
tags:
  - name: recipes
paths:
  /api/recipes:
    get:
      tags: [recipes]
      responses:
        "200": { description: OK, content: { application/json: { schema: { $ref: "#/components/schemas/Recipe" } } } }
        "404": { $ref: "#/components/schemas/ErrorBody" }
schemas:
  Recipe:
    type: object
`

func TestMergePlacesEachPartInItsBlock(t *testing.T) {
	doc, err := run(t, map[string]string{"recipes": recipes, "notes": "tags:\n  - name: notes\npaths:\n  /api/notes:\n    get:\n      tags: [notes]\n"})
	if err != nil {
		t.Fatal(err)
	}
	order := []string{"- name: platform", "- name: notes", "- name: recipes", "paths:", "/healthz:", "/api/notes:", "/api/recipes:", "components:", "    ErrorBody:", "    Recipe:"}
	last := -1
	for _, want := range order {
		i := strings.Index(doc, want)
		if i < 0 || i < last {
			t.Fatalf("%q is missing or out of order (fragments are merged by file name):\n%s", want, doc)
		}
		last = i
	}
	if !strings.Contains(doc, "\n    Recipe:\n      type: object\n") {
		t.Errorf("schemas must be indented into components.schemas:\n%s", doc)
	}
}

func TestMergeWithoutFragments(t *testing.T) {
	doc, err := run(t, nil)
	if err != nil || doc != base {
		t.Fatalf("a base without fragments stays unchanged: %v\n%s", err, doc)
	}
}

func TestValidationFindsMistakes(t *testing.T) {
	for name, tc := range map[string]struct {
		frags map[string]string
		want  string
	}{
		"path defined twice": {map[string]string{"a": "paths:\n  /x:\n    get: {}\n", "b": "paths:\n  /x:\n    get: {}\n"}, "path defined twice: /x"},
		"path also in base":  {map[string]string{"a": "paths:\n  /healthz:\n    get: {}\n"}, "path defined twice: /healthz"},
		"schema twice":       {map[string]string{"a": "schemas:\n  Dup:\n    type: object\n", "b": "schemas:\n  Dup:\n    type: object\n"}, "schemas/Dup defined twice"},
		"undeclared tag":     {map[string]string{"a": "paths:\n  /x:\n    get:\n      tags: [ghost]\n"}, "tag used but not declared: ghost"},
		"tag declared twice": {map[string]string{"a": "tags:\n  - name: platform\n"}, "tag declared twice: platform"},
		"unresolved ref":     {map[string]string{"a": "paths:\n  /x:\n    get:\n      r: { $ref: \"#/components/schemas/Missing\" }\n"}, "unresolved $ref: schemas/Missing"},
		"unknown key":        {map[string]string{"a": "components:\n  x: 1\n"}, `unknown top-level key "components"`},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := run(t, tc.frags); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want an error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestBaseMustEndWithSchemas(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "base.yaml", strings.Replace(base, "components:\n", "components:\n  parameters:\n    id: { name: id }\n", 1)+"  examples:\n    e: {}\n")
	mods := filepath.Join(dir, "m")
	_ = os.Mkdir(mods, 0o750)
	if _, err := build(filepath.Join(dir, "base.yaml"), mods); err == nil || !strings.Contains(err.Error(), "last block") {
		t.Fatalf("got %v", err)
	}
}
