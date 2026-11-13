package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/gamedeals"
)

// The gamedeals module registers itself; no shared file lists it (#193).
func init() {
	registerModule("gamedeals", func(d moduleDeps) modules.Module { return gamedeals.New(d.db, d.out.client, d.clock) })
}
