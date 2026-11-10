package rss

import (
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
	uc       usecase.Reader
	handlers *controller.Handlers
}

// New builds the feed reader module.
func New(db *sql.DB, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewReader(usecase.Deps{Repo: gateway.NewRepository(db), Fetcher: gateway.NewHTTPFetcher(client), Clock: clock, IDs: ids})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
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
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "rss.refresh", Interval: refreshInterval, Run: m.uc.RefreshAll}}
}
