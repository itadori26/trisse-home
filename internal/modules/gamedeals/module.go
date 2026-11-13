package gamedeals

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/gamedeals/adapter/controller"
	"trisse-home/internal/modules/gamedeals/adapter/gateway"
	"trisse-home/internal/modules/gamedeals/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the game deals rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.GameDeals
	handlers *controller.Handlers
}

// New builds the module (#62).
func New(db *sql.DB, client *http.Client, clock kernel.Clock) *Module {
	uc := usecase.NewGameDeals(usecase.Deps{
		Repo: gateway.NewSettingsRepository(db), Clock: clock,
		Wishlist: gateway.NewSteam(client, gateway.DefaultSteamAPIURL), Free: gateway.NewEpic(client, gateway.DefaultEpicURL),
	})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "gamedeals" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"gamedeals"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/gamedeals/settings", m.handlers.Settings)
	r.Protected("PUT /api/gamedeals/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/gamedeals/overview", m.handlers.Overview)
}

// Jobs implements modules.Module; nothing is requested until a source is enabled.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "gamedeals.refresh", Interval: usecase.RefreshInterval, Run: m.uc.Refresh}}
}
