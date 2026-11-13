package media

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/media/adapter/controller"
	"trisse-home/internal/modules/media/adapter/gateway"
	"trisse-home/internal/modules/media/domain"
	"trisse-home/internal/modules/media/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the media server rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the media module (#61); cipher encrypts the stored token.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	providers := map[domain.Provider]usecase.MediaServerProvider{
		domain.ProviderJellyfin: gateway.NewJellyfin(client),
		domain.ProviderPlex:     gateway.NewPlex(client),
	}
	uc := usecase.NewMedia(gateway.NewSettingsRepository(db, cipher), providers, clock)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "media" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"media"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/media/settings", m.handlers.Settings)
	r.Protected("PUT /api/media/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/media/overview", m.handlers.Overview)
	r.Protected("GET /api/media/posters/{id}", m.handlers.Poster)
}
