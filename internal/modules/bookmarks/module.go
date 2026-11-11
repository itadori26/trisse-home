package bookmarks

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/bookmarks/adapter/controller"
	"trisse-home/internal/modules/bookmarks/adapter/gateway"
	"trisse-home/internal/modules/bookmarks/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the bookmarks rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	seed     usecase.SeedSampleData
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

// New builds the bookmarks module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	d := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids}
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, seed: usecase.NewSeedSampleData(d), handlers: controller.NewHandlers(controller.UseCases{
		List: usecase.NewListBookmarks(d), Create: usecase.NewCreateBookmark(d), Update: usecase.NewUpdateBookmark(d),
		Delete: usecase.NewDeleteBookmark(d), CreateFolder: usecase.NewCreateFolder(d),
		RenameFolder: usecase.NewRenameFolder(d), DeleteFolder: usecase.NewDeleteFolder(d),
		Import: usecase.NewImportBookmarks(d),
	})}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "bookmarks" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"bookmarks"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/bookmarks", h.List)
	r.Protected("POST /api/bookmarks", h.Create)
	r.Protected("PUT /api/bookmarks/{id}", h.Update)
	r.Protected("DELETE /api/bookmarks/{id}", h.Delete)
	r.Protected("POST /api/bookmarks/folders", h.CreateFolder)
	r.Protected("PATCH /api/bookmarks/folders/{id}", h.RenameFolder)
	r.Protected("DELETE /api/bookmarks/folders/{id}", h.DeleteFolder)
	r.Protected("POST /api/bookmarks/import", h.Import)
}
