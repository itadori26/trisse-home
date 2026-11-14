package holidays

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/holidays/adapter/controller"
	"trisse-home/internal/modules/holidays/adapter/gateway"
	"trisse-home/internal/modules/holidays/usecase"
	"trisse-home/internal/shared/kernel"
)

// refreshInterval: the job checks daily; school holidays are reloaded weekly.
const refreshInterval = 24 * time.Hour

// Module wires the holiday rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.Holidays
	handlers *controller.Handlers
}

// New builds the module (#55). Public holidays need no request; school
// holidays are only loaded once an ICS link is configured.
func New(db *sql.DB, client *http.Client, clock kernel.Clock, loc *time.Location) *Module {
	store := gateway.NewStore(db)
	uc := usecase.New(store, gateway.NewICSSource(client, loc), gateway.NewCache(store), clock)
	return &Module{
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		uc:           uc,
		handlers:     controller.NewHandlers(uc),
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "holidays" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"holidays"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/holidays", m.handlers.Upcoming)
	r.Protected("GET /api/holidays/settings", m.handlers.Settings)
	r.Protected("PUT /api/holidays/settings", m.handlers.SaveSettings)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "holidays.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
