package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/arr"
)

func init() {
	registerModule("arr", func(d moduleDeps) modules.Module {
		return arr.New(d.db, d.out.secrets, d.out.integration, d.connections, d.clock, d.ids)
	})
}
