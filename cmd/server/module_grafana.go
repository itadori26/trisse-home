package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/grafana"
)

// The Grafana module registers itself; no shared file lists it (#193).
func init() {
	registerModule("grafana", func(d moduleDeps) modules.Module {
		return grafana.New(d.db, d.out.secrets, d.out.integration, d.connections, d.clock)
	})
}
