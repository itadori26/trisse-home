package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/media"
)

func init() {
	registerModule("media", func(d moduleDeps) modules.Module {
		return media.New(d.db, d.out.secrets, d.out.integration, d.connections, d.clock)
	})
}
