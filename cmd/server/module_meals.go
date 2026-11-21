package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/meals"
)

// The Meal plan module registers itself; no shared file lists it (#193).
func init() {
	registerModule("meals", func(d moduleDeps) modules.Module { return meals.New(d.db, d.clock, d.ids, time.Local) })
}
