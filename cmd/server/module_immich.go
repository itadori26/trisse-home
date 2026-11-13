package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/immich"
)

// The Immich module registers itself; no shared file lists it (#193).
func init() {
	registerModule("immich", func(d moduleDeps) modules.Module {
		return immich.New(d.db, d.out.secrets, d.out.integration, d.clock)
	})
}
