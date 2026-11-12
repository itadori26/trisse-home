package habits

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/habits/adapter/controller"
	"trisse-home/internal/modules/habits/adapter/gateway"
	"trisse-home/internal/modules/habits/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the habit tracker rings and implements modules.Module (#106).
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the habit tracker module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewHabits(usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "habits" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"habits"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/habits", m.handlers.List)
	r.Protected("POST /api/habits", m.handlers.Create)
	r.Protected("PUT /api/habits/{id}", m.handlers.Update)
	r.Protected("DELETE /api/habits/{id}", m.handlers.Delete)
	r.Protected("POST /api/habits/{id}/toggle", m.handlers.Toggle)
}
