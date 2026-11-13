package readlater

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/readlater/adapter/controller"
	"trisse-home/internal/modules/readlater/adapter/gateway"
	"trisse-home/internal/modules/readlater/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the read-later rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the module (#99). Pages and images are only requested when the
// user saves an article; client must guard against internal addresses.
func New(db *sql.DB, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	f := gateway.NewFetcher(client)
	uc := usecase.New(gateway.NewRepository(db), f, f, clock, ids)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "readlater" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"readlater"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/readlater", m.handlers.List)
	r.Protected("POST /api/readlater", m.handlers.Save)
	r.Protected("GET /api/readlater/{id}", m.handlers.Get)
	r.Protected("PUT /api/readlater/{id}/status", m.handlers.SetStatus)
	r.Protected("DELETE /api/readlater/{id}", m.handlers.Delete)
	r.Protected("GET /api/readlater/{id}/images/{n}", m.handlers.Image)
}
