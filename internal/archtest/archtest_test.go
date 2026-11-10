package archtest

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePath = "trisse-home"

// ring classifies a package directory relative to the repository root.
type ring int

const (
	ringOther ring = iota
	ringKernel
	ringDomain
	ringUseCase
	ringAdapter
	ringModuleRoot
)

type pkgInfo struct {
	ring   ring
	module string // feature module name, empty outside internal/modules/<name>
}

func classify(importPath string) pkgInfo {
	rel := strings.TrimPrefix(importPath, modulePath+"/")
	if rel == "internal/shared/kernel" {
		return pkgInfo{ring: ringKernel}
	}
	parts := strings.Split(rel, "/")
	if len(parts) < 3 || parts[0] != "internal" || parts[1] != "modules" {
		return pkgInfo{ring: ringOther}
	}
	info := pkgInfo{module: parts[2], ring: ringModuleRoot}
	if len(parts) > 3 {
		switch parts[3] {
		case "domain":
			info.ring = ringDomain
		case "usecase":
			info.ring = ringUseCase
		case "adapter":
			info.ring = ringAdapter
		}
	}
	return info
}

func isStdlib(path string) bool {
	first, _, _ := strings.Cut(path, "/")
	return !strings.Contains(first, ".") && first != modulePath
}

// allowed reports whether a non-test file of pkg may import imp.
func allowed(pkg pkgInfo, imp string) (bool, string) {
	if isStdlib(imp) {
		return true, ""
	}
	if !strings.HasPrefix(imp, modulePath+"/") {
		// Third-party code is a framework detail and must stay outside the
		// two innermost rings.
		if pkg.ring == ringDomain || pkg.ring == ringUseCase || pkg.ring == ringKernel {
			return false, "inner rings may only use the standard library"
		}
		return true, ""
	}
	target := classify(imp)
	if target.module != "" && pkg.module != "" && target.module != pkg.module {
		return false, "modules must not import each other"
	}
	if pkg.module != "" && strings.HasPrefix(imp, modulePath+"/internal/platform") && pkg.ring != ringModuleRoot {
		return false, "only module.go may receive platform services"
	}
	switch pkg.ring {
	case ringKernel:
		return false, "the shared kernel must not import project packages"
	case ringDomain:
		return target.ring == ringKernel || target.ring == ringDomain, "domain may only import the kernel"
	case ringUseCase:
		return target.ring <= ringUseCase && target.ring != ringOther, "use cases may only import domain and kernel"
	case ringAdapter:
		return target.ring != ringModuleRoot, "adapters must not import the module wiring"
	}
	return true, ""
}

func TestDependencyRule(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	err = filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		relDir, _ := filepath.Rel(root, filepath.Dir(path))
		pkg := classify(modulePath + "/" + filepath.ToSlash(relDir))
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range f.Imports {
			imp := strings.Trim(spec.Path.Value, `"`)
			if ok, why := allowed(pkg, imp); !ok {
				t.Errorf("%s imports %s: %s", strings.TrimPrefix(path, root+string(os.PathSeparator)), imp, why)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllowedRules(t *testing.T) {
	cases := []struct {
		pkg, imp string
		want     bool
	}{
		{"trisse-home/internal/modules/notes/domain", "trisse-home/internal/shared/kernel", true},
		{"trisse-home/internal/modules/notes/domain", "net/http", true}, // depguard forbids specific stdlib packages
		{"trisse-home/internal/modules/notes/domain", "trisse-home/internal/modules/notes/usecase", false},
		{"trisse-home/internal/modules/notes/usecase", "trisse-home/internal/modules/notes/domain", true},
		{"trisse-home/internal/modules/notes/usecase", "trisse-home/internal/modules/todos/domain", false},
		{"trisse-home/internal/modules/notes/usecase", "github.com/x/y", false},
		{"trisse-home/internal/modules/notes/adapter/gateway", "trisse-home/internal/platform/database", false},
		{"trisse-home/internal/modules/notes/adapter/gateway", "trisse-home/internal/modules/notes/usecase", true},
		{"trisse-home/internal/modules/notes", "trisse-home/internal/platform/httpclient", true},
		{"trisse-home/internal/shared/kernel", "trisse-home/internal/shared/httpapi", false},
	}
	for _, c := range cases {
		if got, _ := allowed(classify(c.pkg), c.imp); got != c.want {
			t.Errorf("%s -> %s: got %v, want %v", c.pkg, c.imp, got, c.want)
		}
	}
}
