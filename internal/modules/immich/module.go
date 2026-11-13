package immich

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/immich/adapter/controller"
	"trisse-home/internal/modules/immich/adapter/gateway"
	"trisse-home/internal/modules/immich/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the Immich rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the Immich module (#85); cipher encrypts the API key.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	uc := usecase.NewImmich(gateway.NewSettingsRepository(db, cipher), gateway.NewAPI(client), clock)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "immich" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"immich"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/immich/settings", m.handlers.Settings)
	r.Protected("PUT /api/immich/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/immich/overview", m.handlers.Overview)
	r.Protected("GET /api/immich/thumbnails/{id}", m.handlers.Thumbnail)
}
