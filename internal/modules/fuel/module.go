package fuel

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/fuel/adapter/controller"
	"trisse-home/internal/modules/fuel/adapter/gateway"
	"trisse-home/internal/modules/fuel/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the fuel price rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the fuel module (#53); apiURL is the Tankerkönig base URL.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, apiURL string, clock kernel.Clock) *Module {
	if apiURL == "" {
		apiURL = gateway.DefaultTankerkoenigURL
	}
	uc := usecase.NewFuel(gateway.NewSettingsRepository(db, cipher), gateway.NewTankerkoenig(client, apiURL), clock)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "fuel" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"fuel"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/fuel/settings", m.handlers.Settings)
	r.Protected("PUT /api/fuel/settings", m.handlers.SaveSettings)
	r.Protected("PUT /api/fuel/favourites", m.handlers.Favourites)
	r.Protected("GET /api/fuel/overview", m.handlers.Overview)
}
