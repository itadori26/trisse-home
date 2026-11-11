package nextcloud

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/nextcloud/adapter/controller"
	"trisse-home/internal/modules/nextcloud/adapter/gateway"
	"trisse-home/internal/modules/nextcloud/usecase"
	"trisse-home/internal/shared/kernel"
)

const refreshInterval = 5 * time.Minute

// Module wires the Nextcloud rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Nextcloud
	handlers *controller.Handlers
}

// New builds the Nextcloud module; cipher encrypts the app password.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	uc := usecase.NewNextcloud(gateway.NewSettingsRepository(db, cipher), gateway.NewAPI(client), clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "nextcloud" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"nextcloud"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/nextcloud/settings", m.handlers.Settings)
	r.Protected("PUT /api/nextcloud/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/nextcloud/status", m.handlers.Status)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "nextcloud.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
