package waste

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/waste/adapter/controller"
	"trisse-home/internal/modules/waste/adapter/gateway"
	"trisse-home/internal/modules/waste/usecase"
	"trisse-home/internal/shared/kernel"
)

// refreshInterval: collection calendars rarely change.
const refreshInterval = 6 * time.Hour

// Module wires the waste rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.Waste
	handlers *controller.Handlers
}

// New builds the waste module; floating ICS times are interpreted in loc.
func New(db *sql.DB, client *http.Client, clock kernel.Clock, loc *time.Location) *Module {
	uc := usecase.NewWaste(gateway.NewSettingsRepository(db), gateway.NewICSSource(client, loc), clock)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "waste" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"waste"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/waste/settings", m.handlers.Settings)
	r.Protected("PUT /api/waste/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/waste/upcoming", m.handlers.Upcoming)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "waste.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
