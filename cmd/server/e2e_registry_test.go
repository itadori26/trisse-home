package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/platform/config"
)

// A module adds its end-to-end check from its own cmd/server/api_<name>_test.go
// instead of editing api_test.go (#194): a fake upstream, optional config
// overrides and the checks that run against the logged-in test server.

// e2eEnv is what a module check gets: the logged-in client and the fake upstream.
type e2eEnv struct {
	t  *testing.T
	c  *client
	up *httptest.Server
}

// moduleE2E describes the end-to-end test of one module.
type moduleE2E struct {
	// Name of the module, used as the subtest name.
	Name string
	// Upstream registers the handlers of the module's fake external service
	// (use a path prefix of your own, e.g. /immich/…).
	Upstream func(mux *http.ServeMux)
	// Config points the module at the fake upstream (e.g. an API base URL).
	Config func(cfg *config.Config, upstream string)
	// Check runs against the logged-in server after the shared checks.
	Check func(e e2eEnv)
}

var moduleE2ESpecs []moduleE2E

// registerModuleE2E is called from init() in cmd/server/api_<name>_test.go.
func registerModuleE2E(m moduleE2E) { moduleE2ESpecs = append(moduleE2ESpecs, m) }

func registerModuleUpstreams(mux *http.ServeMux) {
	for _, m := range moduleE2ESpecs {
		if m.Upstream != nil {
			m.Upstream(mux)
		}
	}
}

func applyModuleE2EConfig(cfg *config.Config, upstream string) {
	for _, m := range moduleE2ESpecs {
		if m.Config != nil {
			m.Config(cfg, upstream)
		}
	}
}

// runModuleE2E runs the registered checks as subtests with their own client view.
func runModuleE2E(t *testing.T, c *client, up *httptest.Server) {
	t.Helper()
	for _, m := range moduleE2ESpecs {
		if m.Check == nil {
			continue
		}
		t.Run(m.Name, func(t *testing.T) {
			view := *c // same session and CSRF token, failures reported on the subtest
			view.t = t
			m.Check(e2eEnv{t: t, c: &view, up: up})
		})
	}
}
