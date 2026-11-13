package main

import (
	"testing"

	"trisse-home/internal/modules"
	"trisse-home/internal/platform/config"
)

func TestRegisterModuleSortsAndRejectsDuplicates(t *testing.T) {
	saved := moduleFactories
	t.Cleanup(func() { moduleFactories = saved })
	moduleFactories = nil

	build := func(moduleDeps) modules.Module { return nil }
	registerModule("zeta", build)
	registerModule("alpha", build)
	if got := []string{moduleFactories[0].name, moduleFactories[1].name}; got[0] != "alpha" || got[1] != "zeta" {
		t.Fatalf("modules are built in alphabetical order, got %v", got)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("registering a module twice must panic")
		}
	}()
	registerModule("alpha", build)
}

func TestSelfRegisteredModulesRespectDisabledList(t *testing.T) {
	saved := moduleFactories
	t.Cleanup(func() { moduleFactories = saved })
	moduleFactories = nil

	built := map[string]bool{}
	for _, name := range []string{"one", "two"} {
		registerModule(name, func(moduleDeps) modules.Module { built[name] = true; return nil })
	}
	// A nil module must not be registered, so use a registry that is never
	// asked: disable both and check nothing is built.
	cfg := config.Config{DisabledModules: []string{"one", "two"}}
	registerSelfRegistered(moduleDeps{cfg: cfg, registry: &modules.Registry{}})
	if len(built) != 0 {
		t.Fatalf("disabled modules were built: %v", built)
	}
}
