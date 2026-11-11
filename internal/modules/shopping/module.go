package shopping

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/shopping/adapter/controller"
	"trisse-home/internal/modules/shopping/adapter/gateway"
	"trisse-home/internal/modules/shopping/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the shopping list rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the shopping list module; cipher encrypts the app password.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, ids kernel.IDGenerator) *Module {
	uc := usecase.NewShopping(gateway.NewSettingsRepository(db, cipher), gateway.NewCalDAVStore(client), ids)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "shopping" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"shopping"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/shopping/settings", h.Settings)
	r.Protected("PUT /api/shopping/settings", h.SaveSettings)
	r.Protected("GET /api/shopping/items", h.List)
	r.Protected("POST /api/shopping/items", h.Add)
	r.Protected("PUT /api/shopping/items/{id}", h.SetDone)
	r.Protected("DELETE /api/shopping/items/{id}", h.Delete)
	r.Protected("POST /api/shopping/clear-completed", h.ClearCompleted)
}
