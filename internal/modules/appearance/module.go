package appearance

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/appearance/adapter/controller"
	"trisse-home/internal/modules/appearance/adapter/gateway"
	"trisse-home/internal/modules/appearance/usecase"
)

// Module wires the appearance rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	themes   *controller.ThemeHandlers
}

// New builds the appearance module.
func New(db *sql.DB) *Module {
	repo := gateway.NewRepository(db)
	themes := usecase.NewThemes(repo, gateway.NewThemeStore(db), gateway.Codec{}, gateway.BuiltinThemes())
	return &Module{
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(usecase.NewGetAppearance(repo), usecase.NewUpdateAppearance(repo)),
		themes:       controller.NewThemeHandlers(themes),
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "appearance" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/appearance", m.handlers.Get)
	r.Protected("PUT /api/appearance", m.handlers.Update)
	r.Protected("GET /api/appearance/themes", m.themes.Gallery)
	r.Protected("POST /api/appearance/themes", m.themes.Import)
	r.Protected("POST /api/appearance/themes/restore", m.themes.Restore)
	r.Protected("GET /api/appearance/themes/{id}", m.themes.Export)
	r.Protected("DELETE /api/appearance/themes/{id}", m.themes.Delete)
	r.Protected("POST /api/appearance/themes/{id}/apply", m.themes.Apply)
}
