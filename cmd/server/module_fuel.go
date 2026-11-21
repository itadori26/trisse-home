package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/fuel"
)

// The fuel module registers itself; no shared file lists it (#193).
func init() {
	registerModule("fuel", func(d moduleDeps) modules.Module {
		return fuel.New(d.db, d.out.secrets, d.out.client, d.cfg.FuelAPIURL, d.connections, d.clock)
	})
}
