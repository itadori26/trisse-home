package focus

import (
	"database/sql"
	"io/fs"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/focus/adapter/controller"
	"trisse-home/internal/modules/focus/adapter/gateway"
	"trisse-home/internal/modules/focus/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the focus timer rings and implements modules.Module (#58).
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	uc       usecase.Focus
}

// New builds the focus timer module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewFocus(usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc), uc: uc}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "focus" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"focus"} }

// Jobs implements modules.Module: old sessions are pruned once a day.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "focus.prune", Interval: 24 * time.Hour, Run: m.uc.Prune}}
}

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/focus/stats", m.handlers.Stats)
	r.Protected("POST /api/focus/sessions", m.handlers.Record)
}
