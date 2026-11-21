package meals

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/meals/adapter/controller"
	"waylight/internal/modules/meals/adapter/gateway"
	"waylight/internal/modules/meals/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the meal plan rings and implements modules.Module (#682).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	uc       usecase.Meals
	clock    kernel.Clock
}

// New builds the meal plan module; loc decides what "today" is.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, loc *time.Location) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids, Location: loc}
	uc := usecase.NewMeals(deps)
	return &Module{
		seedDemo:     func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) },
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(uc),
		uc:           uc,
		clock:        clock,
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "meals" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"meals"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/meals", m.handlers.List)
	r.Protected("POST /api/meals", m.handlers.Create)
	r.Protected("PUT /api/meals/{id}", m.handlers.Update)
	r.Protected("DELETE /api/meals/{id}", m.handlers.Delete)
}

// CompactState implements modules.CompactProvider (#597): what is on the
// table today.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	today, err := usecase.Today(ctx, m.uc)
	if err != nil {
		return kernel.CompactState{}, err
	}
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: m.clock.Now()}
	switch len(today) {
	case 0:
		state.Title = kernel.NewText("Nothing planned for today")
		state.Empty = true
	case 1:
		state.Title = kernel.NewTextf("Today: {dish}", "dish", today[0].Dish)
	default:
		// The last dish of the day is the main one in a plan with
		// breakfast and lunch; the others go into the detail line.
		state.Title = kernel.NewTextf("Today: {dish}", "dish", today[len(today)-1].Dish)
		state.Detail = kernel.NewTextf("and {count} more", "count", len(today)-1)
	}
	return state, nil
}
