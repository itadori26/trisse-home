package notes

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/notes/adapter/controller"
	"trisse-home/internal/modules/notes/adapter/gateway"
	"trisse-home/internal/modules/notes/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the notes rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	seed     usecase.SeedSampleData
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

// New builds the notes module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	repo := gateway.NewRepository(db)
	d := usecase.Deps{Repo: repo, Clock: clock, IDs: ids, Search: repo, Revisions: repo}
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, seed: usecase.NewSeedSampleData(d), handlers: controller.NewHandlers(controller.UseCases{
		List: usecase.NewListNotes(d), Get: usecase.NewGetNote(d), Save: usecase.NewSaveNote(d), Delete: usecase.NewDeleteNote(d),
		Search: usecase.NewSearchNotes(d), Backlinks: usecase.NewListBacklinks(d), Revisions: usecase.NewListRevisions(d),
		Revision: usecase.NewGetRevision(d), Restore: usecase.NewRestoreRevision(d),
	})}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "notes" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"notes"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/notes", h.List)
	r.Protected("GET /api/notes/search", h.Search)
	r.Protected("GET /api/notes/{id}/backlinks", h.Backlinks)
	r.Protected("GET /api/notes/{id}/revisions", h.Revisions)
	r.Protected("GET /api/notes/{id}/revisions/{rev}", h.Revision)
	r.Protected("POST /api/notes/{id}/revisions/{rev}/restore", h.Restore)
	r.Protected("POST /api/notes", h.Create)
	r.Protected("GET /api/notes/{id}", h.Get)
	r.Protected("PUT /api/notes/{id}", h.Update)
	r.Protected("DELETE /api/notes/{id}", h.Delete)
}
