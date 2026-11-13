package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestBuildAddsConventionalModulesOnly(t *testing.T) {
	dir := t.TempDir()
	mods := filepath.Join(dir, "modules")
	// media follows the convention; irregular is listed explicitly and must
	// not be added a second time; plain has a controller without dto.go.
	write(t, filepath.Join(mods, "media/adapter/presenter/presenter.go"))
	write(t, filepath.Join(mods, "media/adapter/controller/dto.go"))
	write(t, filepath.Join(mods, "irregular/adapter/controller/dto.go"))
	write(t, filepath.Join(mods, "plain/adapter/controller/handlers.go"))
	write(t, filepath.Join(mods, "onlytests/adapter/presenter/presenter_test.go"))

	base := filepath.Join(dir, "tygo.yaml")
	body := "packages:\n" +
		"  - path: \"trisse-home/internal/modules/irregular/adapter/controller\"\n" +
		"    output_path: \"web/src/shared/api-client/generated/irregular.ts\"\n"
	if err := os.WriteFile(base, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := build(base, mods)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`path: "trisse-home/internal/modules/media/adapter/presenter"`,
		`output_path: "web/src/shared/api-client/generated/media.ts"`,
		`path: "trisse-home/internal/modules/media/adapter/controller"`,
		`output_path: "web/src/shared/api-client/generated/media-requests.ts"`,
		`include_files: ["dto.go"]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in:\n%s", want, got)
		}
	}
	if n := strings.Count(got, "modules/irregular/"); n != 1 {
		t.Errorf("an explicitly listed package must not be added again, found %d times", n)
	}
	for _, unwanted := range []string{"plain", "onlytests"} {
		if strings.Contains(got, "modules/"+unwanted+"/") {
			t.Errorf("%s must not get an entry:\n%s", unwanted, got)
		}
	}
}
