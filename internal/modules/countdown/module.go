package countdown

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/countdown/adapter/controller"
	"trisse-home/internal/modules/countdown/adapter/gateway"
	"trisse-home/internal/modules/countdown/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the countdown rings and implements modules.Module (#54).
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the countdown module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewCountdowns(usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "countdown" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"countdown"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/countdowns", m.handlers.List)
	r.Protected("POST /api/countdowns", m.handlers.Create)
	r.Protected("PUT /api/countdowns/{id}", m.handlers.Update)
	r.Protected("DELETE /api/countdowns/{id}", m.handlers.Delete)
}
