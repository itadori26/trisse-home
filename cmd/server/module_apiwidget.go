package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/apiwidget"
	"trisse-home/internal/modules/apiwidget/adapter/gateway"
	"trisse-home/internal/platform/httpclient"
)

// The JSON widget registers itself; no shared file lists it (#193).
func init() {
	registerModule("apiwidget", func(d moduleDeps) modules.Module {
		fetcher := gateway.Fetcher{
			Private: httpclient.New(httpclient.WithPrivateOnly()),
			Public:  httpclient.New(httpclient.WithIntegrationGuard()),
		}
		if d.cfg.DemoMode {
			fetcher = gateway.Fetcher{Private: d.out.integration, Public: d.out.integration}
		}
		return apiwidget.New(apiwidget.Deps{DB: d.db, Cipher: d.out.secrets, Fetcher: fetcher, Clock: d.clock, IDs: d.ids})
	})
}
