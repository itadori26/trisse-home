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
	handlers *controller.Handlers
	seed     usecase.SeedSampleData
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

// New builds the notes module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	d := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids}
	return &Module{seed: usecase.NewSeedSampleData(d), handlers: controller.NewHandlers(controller.UseCases{
		List: usecase.NewListNotes(d), Get: usecase.NewGetNote(d), Save: usecase.NewSaveNote(d), Delete: usecase.NewDeleteNote(d),
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
	r.Protected("POST /api/notes", h.Create)
	r.Protected("GET /api/notes/{id}", h.Get)
	r.Protected("PUT /api/notes/{id}", h.Update)
	r.Protected("DELETE /api/notes/{id}", h.Delete)
}
