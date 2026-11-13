package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/habits"
)

// The Habits module registers itself; no shared file lists it (#193).
func init() {
	registerModule("habits", func(d moduleDeps) modules.Module { return habits.New(d.db, d.clock, d.ids) })
}
