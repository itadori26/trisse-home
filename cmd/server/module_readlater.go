package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/readlater"
)

// The read-later module registers itself; no shared file lists it (#193).
// Pages are fetched with the guarded client (no loopback or link-local
// addresses, checked on every connection and redirect).
func init() {
	registerModule("readlater", func(d moduleDeps) modules.Module {
		client := d.out.guarded
		if client == nil { // demo mode has no guarded client; its transport answers locally
			client = d.out.client
		}
		return readlater.New(d.db, client, d.clock, d.ids)
	})
}
