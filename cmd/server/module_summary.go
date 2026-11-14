package main

import (
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/summary"
	"trisse-home/internal/platform/httpclient"
)

// ollamaTimeout leaves a local model on a CPU time to answer (#75).
const ollamaTimeout = 3 * time.Minute

// The summary module registers itself; no shared file lists it (#193).
func init() {
	registerModule("summary", func(d moduleDeps) modules.Module {
		model := d.out.integration
		if !d.cfg.DemoMode {
			model = httpclient.New(httpclient.WithIntegrationGuard(), httpclient.WithTimeout(ollamaTimeout))
		}
		pages := d.out.guarded
		if pages == nil { // demo mode has no guarded client
			pages = d.out.client
		}
		return summary.New(summary.Config{DB: d.db, Model: model, Pages: pages, Clock: d.clock})
	})
}
