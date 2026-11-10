package adguard

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/adguard/adapter/controller"
	"trisse-home/internal/modules/adguard/adapter/gateway"
	"trisse-home/internal/modules/adguard/usecase"
	"trisse-home/internal/shared/kernel"
)

const refreshInterval = time.Minute

// Module wires the AdGuard rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.AdGuard
	handlers *controller.Handlers
}

// New builds the AdGuard module; cipher encrypts the stored password.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	uc := usecase.NewAdGuard(gateway.NewSettingsRepository(db, cipher), gateway.NewAPI(client), clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "adguard" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"adguard"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/adguard/settings", m.handlers.Settings)
	r.Protected("PUT /api/adguard/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/adguard/stats", m.handlers.Stats)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "adguard.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
