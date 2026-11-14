package drop

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/drop/adapter/controller"
	"trisse-home/internal/modules/drop/adapter/gateway"
	"trisse-home/internal/modules/drop/usecase"
	"trisse-home/internal/shared/kernel"
)

// purgeInterval: expired items are removed within a quarter of an hour (and
// are never listed or served once expired).
const purgeInterval = 15 * time.Minute

// Module wires the drop zone rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Drop
	handlers *controller.Handlers
}

// New builds the module (#101); it makes no outbound requests.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.New(gateway.NewRepository(db), clock, ids)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "drop" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"drop"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/drop", m.handlers.List)
	r.Protected("POST /api/drop/text", m.handlers.AddText)
	r.Protected("POST /api/drop/files", m.handlers.AddFile)
	r.Protected("DELETE /api/drop/{id}", m.handlers.Delete)
	r.Protected("GET /api/drop/{id}/file", m.handlers.File)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "drop.purge", Interval: purgeInterval, Run: func(ctx context.Context) error { return m.uc.PurgeExpired(ctx) }}}
}
