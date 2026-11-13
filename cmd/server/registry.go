package main

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/platform/clock"
	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/idgen"
	"trisse-home/internal/shared/kernel"
)

// moduleDeps is everything a module constructor may need. Modules that live
// in their own cmd/server/module_<name>.go file (see registerModule) receive
// it instead of capturing locals of buildModules (#193).
type moduleDeps struct {
	cfg      config.Config
	db       *sql.DB
	clock    clock.System
	ids      idgen.Random
	out      outbound
	sink     kernel.ObservationSink
	alerts   *alertRelay
	auth     *auth.Module
	registry *modules.Registry
}

type moduleFactory struct {
	name  string
	build func(moduleDeps) modules.Module
}

// moduleFactories holds the modules that register themselves from their own
// file. A new module therefore touches no shared file: `make new-module`
// creates module_<name>.go, which calls registerModule in init().
var moduleFactories []moduleFactory

// registerModule makes a module part of the server. Call it from init() in
// cmd/server/module_<name>.go. Modules are built in alphabetical order after
// the ones listed in buildModules, so none of them may depend on the build
// order of another.
func registerModule(name string, build func(moduleDeps) modules.Module) {
	if strings.TrimSpace(name) == "" || build == nil {
		panic("registerModule: name and constructor are required")
	}
	for _, f := range moduleFactories {
		if f.name == name {
			panic(fmt.Sprintf("registerModule: module %q is registered twice", name))
		}
	}
	moduleFactories = append(moduleFactories, moduleFactory{name: name, build: build})
	slices.SortFunc(moduleFactories, func(a, b moduleFactory) int { return strings.Compare(a.name, b.name) })
}

// registerSelfRegistered builds every enabled self-registering module.
func registerSelfRegistered(deps moduleDeps) {
	for _, f := range moduleFactories {
		if deps.cfg.ModuleEnabled(f.name) {
			deps.registry.Register(f.build(deps))
		}
	}
}
