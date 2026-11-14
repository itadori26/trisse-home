package rss

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/rss/adapter/controller"
	"trisse-home/internal/modules/rss/adapter/gateway"
	"trisse-home/internal/modules/rss/usecase"
	"trisse-home/internal/shared/kernel"
)

const refreshInterval = 30 * time.Minute

// Module wires the feed reader rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc          usecase.Reader
	seed        usecase.SeedSampleData
	sampleFeeds []string
	handlers    *controller.Handlers
}

// sampleRefreshTimeout bounds the first fetch of the example feeds.
const sampleRefreshTimeout = time.Minute

// New builds the feed reader module; sampleFeeds are subscribed on a fresh
// installation.
func New(db *sql.DB, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator, sampleFeeds []string) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Fetcher: gateway.NewHTTPFetcher(client), OPML: gateway.OPML{}, Clock: clock, IDs: ids}
	uc := usecase.NewReader(deps)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, uc: uc, seed: usecase.NewSeedSampleData(deps), sampleFeeds: sampleFeeds, handlers: controller.NewHandlers(uc)}
}

// SeedSampleData implements modules.Seeder. The example feeds are fetched in
// the background so items appear within seconds without delaying setup.
func (m *Module) SeedSampleData(ctx context.Context) error {
	added, err := m.seed.Execute(ctx, m.sampleFeeds)
	if err != nil || !added {
		return err
	}
	go func() {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), sampleRefreshTimeout)
		defer cancel()
		_ = m.uc.RefreshAll(bg) // failures are stored on the feeds and retried by the job
	}()
	return nil
}

// Name implements modules.Module.
func (m *Module) Name() string { return "rss" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"rss"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/rss/feeds", h.Feeds)
	r.Protected("POST /api/rss/feeds", h.Subscribe)
	r.Protected("DELETE /api/rss/feeds/{id}", h.Unsubscribe)
	r.Protected("GET /api/rss/items", h.Items)
	r.Protected("PUT /api/rss/items/{id}/read", h.SetRead)
	r.Protected("POST /api/rss/read-all", h.MarkAllRead)
	r.Protected("POST /api/rss/refresh", h.Refresh)
	r.Protected("GET /api/rss/opml", h.ExportOPML)
	r.Protected("POST /api/rss/opml", h.ImportOPML)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "rss.refresh", Interval: refreshInterval, Run: m.uc.RefreshAll}}
}
