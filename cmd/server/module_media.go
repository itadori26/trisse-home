package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/media"
)

func init() {
	registerModule("media", func(d moduleDeps) modules.Module {
		return media.New(d.db, d.out.secrets, d.out.integration, d.clock)
	})
}
