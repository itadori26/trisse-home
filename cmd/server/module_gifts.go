package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/gifts"
)

// The Gift ideas module registers itself; no shared file lists it (#193).
func init() {
	registerModule("gifts", func(d moduleDeps) modules.Module { return gifts.New(d.db, d.clock, d.ids) })
}
