package speedtest

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/speedtest/adapter/controller"
	"trisse-home/internal/modules/speedtest/adapter/gateway"
	"trisse-home/internal/modules/speedtest/usecase"
	"trisse-home/internal/shared/kernel"
)

// refreshInterval: Speedtest Tracker usually tests hourly.
const refreshInterval = 10 * time.Minute

// Module wires the speedtest rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Speedtest
	handlers *controller.Handlers
}

// New builds the speedtest module; cipher encrypts the stored token and
// sink receives threshold observations.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	uc := usecase.NewSpeedtest(gateway.NewSettingsRepository(db, cipher), gateway.NewAPI(client), sink, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "speedtest" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"speedtest"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/speedtest/settings", m.handlers.Settings)
	r.Protected("PUT /api/speedtest/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/speedtest/results", m.handlers.Results)
}

// Jobs implements modules.Module; it does nothing until configured.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "speedtest.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
