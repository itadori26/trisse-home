package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/fuel"
)

// The fuel module registers itself; no shared file lists it (#193).
func init() {
	registerModule("fuel", func(d moduleDeps) modules.Module {
		return fuel.New(d.db, d.out.secrets, d.out.client, d.cfg.FuelAPIURL, d.clock)
	})
}
