// Command openapi builds the OpenAPI document from api/openapi.base.yaml and
// the module fragments api/modules/*.yaml (#194), so a module documents its
// endpoints in a file of its own instead of editing one shared document.
//
// A fragment has up to three top-level keys, written with the indentation they
// have in the merged document (schemas one level less, the tool indents them):
//
//	tags:        # list entries, e.g. "  - name: recipes"
//	paths:       # path items, e.g. "  /api/recipes:"
//	schemas:     # component schemas, e.g. "  Recipe:"
//
// The tool checks the result: no path or schema is defined twice, every tag
// used by an operation is declared, and every $ref resolves.
//
//	go run ./scripts/openapi -o dist/openapi.yaml
//	go run ./scripts/openapi -check
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func main() {
	base := flag.String("base", "api/openapi.base.yaml", "base document")
	dir := flag.String("modules", "api/modules", "directory with the module fragments")
	out := flag.String("o", "", "write the merged document here")
	check := flag.Bool("check", false, "only validate (no output file needed)")
	flag.Parse()
	doc, err := build(*base, *dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "openapi:", err)
		os.Exit(1)
	}
	if *out == "" && !*check {
		fmt.Fprintln(os.Stderr, "openapi: pass -o <file> or -check")
		os.Exit(2)
	}
	if *out != "" {
		if err := os.MkdirAll(filepath.Dir(*out), 0o750); err != nil {
			fmt.Fprintln(os.Stderr, "openapi:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*out, []byte(doc), 0o600); err != nil {
			fmt.Fprintln(os.Stderr, "openapi:", err)
			os.Exit(1)
		}
	}
}

// build merges the base document with the fragments and validates the result.
func build(basePath, dir string) (string, error) {
	baseText, err := os.ReadFile(basePath) //nolint:gosec // path from the command line
	if err != nil {
		return "", err
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return "", err
	}
	slices.Sort(files)
	frags := make([]fragment, 0, len(files))
	for _, f := range files {
		text, err := os.ReadFile(f) //nolint:gosec // fragments of the repository
		if err != nil {
			return "", err
		}
		frag, err := parseFragment(strings.TrimSuffix(filepath.Base(f), ".yaml"), string(text))
		if err != nil {
			return "", fmt.Errorf("%s: %w", f, err)
		}
		frags = append(frags, frag)
	}
	doc, err := merge(string(baseText), frags)
	if err != nil {
		return "", err
	}
	return doc, validate(doc)
}

type fragment struct {
	name                 string
	tags, paths, schemas []string
}

var topKey = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9_]*):\s*$`)

func parseFragment(name, text string) (fragment, error) {
	f := fragment{name: name}
	var cur *[]string
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if m := topKey.FindStringSubmatch(line); m != nil {
			switch m[1] {
			case "tags":
				cur = &f.tags
			case "paths":
				cur = &f.paths
			case "schemas":
				cur = &f.schemas
			default:
				return f, fmt.Errorf("unknown top-level key %q (allowed: tags, paths, schemas)", m[1])
			}
			continue
		}
		if line != "" && line[0] != ' ' && line[0] != '#' {
			return f, fmt.Errorf("unexpected line %q", line)
		}
		if cur != nil {
			*cur = append(*cur, line)
		}
	}
	for _, s := range []*[]string{&f.tags, &f.paths, &f.schemas} {
		*s = trimBlank(*s)
	}
	return f, nil
}

func trimBlank(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func indent(lines []string, by string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		if strings.TrimSpace(l) != "" {
			l = by + l
		}
		out[i] = l
	}
	return out
}

// merge inserts the fragments at the end of the tags, paths and schemas blocks of the base.
func merge(base string, frags []fragment) (string, error) {
	lines := strings.Split(strings.TrimRight(base, "\n"), "\n")
	idx := func(prefix string) int { return slices.IndexFunc(lines, func(l string) bool { return l == prefix }) }
	tagsAt, pathsAt, compAt, schemasAt := idx("tags:"), idx("paths:"), idx("components:"), idx("  schemas:")
	if tagsAt < 0 || pathsAt < tagsAt || compAt < pathsAt || schemasAt < compAt {
		return "", errors.New("base: needs the top-level blocks tags, paths and components in this order, and components.schemas")
	}
	for _, l := range lines[schemasAt+1:] {
		if strings.TrimSpace(l) == "" || strings.HasPrefix(l, "    ") || strings.HasPrefix(strings.TrimSpace(l), "#") {
			continue
		}
		return "", errors.New("base: components.schemas must be the last block of the document")
	}
	var tags, paths, schemas []string
	for _, f := range frags {
		tags = append(tags, f.tags...)
		if len(f.paths) > 0 {
			paths = append(paths, "  # "+strings.Repeat("-", 64)+" "+f.name)
			paths = append(paths, f.paths...)
		}
		schemas = append(schemas, indent(f.schemas, "  ")...)
	}
	insertAfterLastContent := func(lines []string, from, to int, add []string) []string {
		at := to
		for at > from+1 && strings.TrimSpace(lines[at-1]) == "" {
			at--
		}
		return slices.Insert(slices.Clone(lines), at, add...)
	}
	// Insert from the bottom so earlier positions stay valid.
	lines = append(slices.Clone(lines), schemas...)
	lines = insertAfterLastContent(lines, pathsAt, compAt, paths)
	lines = insertAfterLastContent(lines, tagsAt, pathsAt, tags)
	return strings.Join(lines, "\n") + "\n", nil
}

var (
	pathKey   = regexp.MustCompile(`^  (/\S*):\s*$`)
	tagDecl   = regexp.MustCompile(`^  - name: (\S+)\s*$`)
	tagUse    = regexp.MustCompile(`tags:\s*\[([^\]]*)\]`)
	refUse    = regexp.MustCompile(`\$ref:\s*["']#/components/(schemas|responses|parameters)/([^"']+)["']`)
	component = regexp.MustCompile(`^    ([A-Za-z0-9_.-]+):`)
)

// validate checks duplicates, tags and references of the merged document.
func validate(doc string) error {
	var problems []string
	paths, tags := map[string]bool{}, map[string]bool{}
	comps := map[string]map[string]bool{"schemas": {}, "responses": {}, "parameters": {}}
	section, kind := "", ""
	for _, line := range strings.Split(doc, "\n") {
		if line != "" && line[0] != ' ' && line[0] != '#' {
			section, kind = strings.TrimSuffix(line, ":"), ""
			continue
		}
		switch section {
		case "tags":
			if m := tagDecl.FindStringSubmatch(line); m != nil {
				if tags[m[1]] {
					problems = append(problems, "tag declared twice: "+m[1])
				}
				tags[m[1]] = true
			}
		case "paths":
			if m := pathKey.FindStringSubmatch(line); m != nil {
				if paths[m[1]] {
					problems = append(problems, "path defined twice: "+m[1])
				}
				paths[m[1]] = true
			}
		case "components":
			if strings.HasPrefix(line, "  ") && len(line) > 2 && line[2] != ' ' && line[2] != '#' {
				kind = strings.TrimSuffix(strings.TrimSpace(line), ":")
			} else if m := component.FindStringSubmatch(line); m != nil && comps[kind] != nil {
				if comps[kind][m[1]] {
					problems = append(problems, fmt.Sprintf("%s/%s defined twice", kind, m[1]))
				}
				comps[kind][m[1]] = true
			}
		}
	}
	for _, line := range strings.Split(doc, "\n") {
		for _, m := range tagUse.FindAllStringSubmatch(line, -1) {
			for _, t := range strings.Split(m[1], ",") {
				if t = strings.TrimSpace(t); t != "" && !tags[t] {
					problems = append(problems, "tag used but not declared: "+t)
				}
			}
		}
		for _, m := range refUse.FindAllStringSubmatch(line, -1) {
			if !comps[m[1]][m[2]] {
				problems = append(problems, fmt.Sprintf("unresolved $ref: %s/%s", m[1], m[2]))
			}
		}
	}
	slices.Sort(problems)
	problems = slices.Compact(problems)
	if len(problems) > 0 {
		return fmt.Errorf("invalid OpenAPI document:\n  %s", strings.Join(problems, "\n  "))
	}
	return nil
}
