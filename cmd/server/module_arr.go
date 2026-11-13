package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/arr"
)

func init() {
	registerModule("arr", func(d moduleDeps) modules.Module {
		return arr.New(d.db, d.out.secrets, d.out.integration, d.clock, d.ids)
	})
}
