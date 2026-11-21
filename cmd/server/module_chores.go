package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/chores"
)

// The Chores module registers itself; no shared file lists it (#193).
func init() {
	registerModule("chores", func(d moduleDeps) modules.Module { return chores.New(d.db, d.clock, d.ids, time.Local) })
}
