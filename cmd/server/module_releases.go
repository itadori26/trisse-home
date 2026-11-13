package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/releases"
)

// The releases module registers itself; no shared file lists it (#193).
func init() {
	registerModule("releases", func(d moduleDeps) modules.Module { return releases.New(d.db, d.out.client, d.clock) })
}
