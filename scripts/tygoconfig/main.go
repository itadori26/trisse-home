// Command tygoconfig prints the tygo configuration: tygo.yaml (the shared and
// the irregular entries) plus one entry per module directory that follows the
// naming convention, so a new module needs no edit of tygo.yaml (#193):
//
//	internal/modules/<m>/adapter/presenter           -> generated/<m>.ts
//	internal/modules/<m>/adapter/controller/dto.go   -> generated/<m>-requests.ts
//
// An entry that tygo.yaml already lists for the same package wins, which
// keeps irregular modules explicit.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	modulePath = "trisse-home/internal/modules/"
	outputDir  = "web/src/shared/api-client/generated/"
)

var listed = regexp.MustCompile(`(?m)^\s*-\s*path:\s*"` + regexp.QuoteMeta(modulePath) + `([^/"]+)/adapter/(presenter|controller)"`)

func main() {
	base := flag.String("base", "tygo.yaml", "static tygo configuration")
	modules := flag.String("modules", "internal/modules", "directory with the modules")
	out := flag.String("o", "", "output file (default: stdout)")
	flag.Parse()

	cfg, err := build(*base, *modules)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tygoconfig:", err)
		os.Exit(1)
	}
	if *out == "" {
		fmt.Print(cfg)
		return
	}
	if err := os.WriteFile(*out, []byte(cfg), 0o644); err != nil { //nolint:gosec // build artefact, not a secret
		fmt.Fprintln(os.Stderr, "tygoconfig:", err)
		os.Exit(1)
	}
}

func build(basePath, modulesDir string) (string, error) {
	raw, err := os.ReadFile(basePath) //nolint:gosec // path comes from the command line of a developer tool
	if err != nil {
		return "", err
	}
	text := string(raw)
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	have := map[string]bool{}
	for _, m := range listed.FindAllStringSubmatch(text, -1) {
		have[m[1]+"/"+m[2]] = true
	}

	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return "", err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(text)
	for _, name := range names {
		adapter := filepath.Join(modulesDir, name, "adapter")
		if !have[name+"/presenter"] && hasGoFiles(filepath.Join(adapter, "presenter")) {
			fmt.Fprintf(&b, "  - path: %q\n    output_path: %q\n", modulePath+name+"/adapter/presenter", outputDir+name+".ts")
		}
		if !have[name+"/controller"] && exists(filepath.Join(adapter, "controller", "dto.go")) {
			fmt.Fprintf(&b, "  - path: %q\n    output_path: %q\n    include_files: [\"dto.go\"]\n",
				modulePath+name+"/adapter/controller", outputDir+name+"-requests.ts")
		}
	}
	return b.String(), nil
}

func hasGoFiles(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if d.IsDir() && p != dir {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
