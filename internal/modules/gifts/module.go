package gifts

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/gifts/adapter/controller"
	"waylight/internal/modules/gifts/adapter/gateway"
	"waylight/internal/modules/gifts/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the gift idea rings and implements modules.Module (#684).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	uc       usecase.Gifts
	clock    kernel.Clock
}

// New builds the gift idea module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids}
	uc := usecase.NewGifts(deps)
	return &Module{
		seedDemo:     func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) },
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(uc),
		uc:           uc,
		clock:        clock,
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "gifts" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"gifts"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/gifts", m.handlers.List)
	r.Protected("POST /api/gifts", m.handlers.Create)
	r.Protected("PUT /api/gifts/{id}", m.handlers.Update)
	r.Protected("DELETE /api/gifts/{id}", m.handlers.Delete)
}

// CompactState implements modules.CompactProvider (#597): how many ideas
// are still to buy. The state names no person, so a compact tile on a
// shared screen gives nothing away.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	list, err := m.uc.List(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	open := usecase.OpenCount(list)
	state := kernel.CompactState{Level: kernel.CompactNone, Count: kernel.IntPtr(open), UpdatedAt: m.clock.Now()}
	switch open {
	case 0:
		state.Title = kernel.NewText("No open gift ideas")
		state.Count = nil
		state.Empty = true
	case 1:
		state.Title = kernel.NewText("1 gift idea")
	default:
		state.Title = kernel.NewTextf("{count} gift ideas", "count", open)
	}
	return state, nil
}
