package chores

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"net/url"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/chores/adapter/controller"
	"waylight/internal/modules/chores/adapter/gateway"
	"waylight/internal/modules/chores/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the chore plan rings and implements modules.Module (#683).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	uc       usecase.Chores
	clock    kernel.Clock
}

// New builds the chore plan module; loc decides what "today" is.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, loc *time.Location) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids, Location: loc}
	uc := usecase.NewChores(deps)
	return &Module{
		seedDemo:     func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) },
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(uc),
		uc:           uc,
		clock:        clock,
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "chores" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"chores"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/chores", m.handlers.List)
	r.Protected("POST /api/chores", m.handlers.Create)
	r.Protected("PUT /api/chores/{id}", m.handlers.Update)
	r.Protected("DELETE /api/chores/{id}", m.handlers.Delete)
	r.Protected("POST /api/chores/{id}/done", m.handlers.Done)
}

// CompactState implements modules.CompactProvider (#597): how many chores
// are due today, with the first one to tick off.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	list, err := m.uc.List(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	var due []usecase.ChoreView
	overdue := false
	for _, c := range list {
		if c.DueNow {
			due = append(due, c)
			overdue = overdue || c.Overdue
		}
	}
	state := kernel.CompactState{Level: kernel.CompactNone, Count: kernel.IntPtr(len(due)), UpdatedAt: m.clock.Now()}
	switch {
	case len(list) == 0:
		state.Title = kernel.NewText("No chores yet")
		state.Count = nil
		state.Empty = true
		return state, nil
	case len(due) == 0:
		state.Title = kernel.NewText("Nothing to do today")
		state.Level = kernel.CompactOK
		state.Empty = true
		return state, nil
	case len(due) == 1:
		state.Title = kernel.NewText("1 task today")
	default:
		state.Title = kernel.NewTextf("{count} tasks today", "count", len(due))
	}
	if overdue {
		state.Level = kernel.CompactAttention
	}
	first := due[0]
	if who := first.Assignee(); who != "" {
		state.Detail = kernel.NewTextf("{title} ({name})", "title", first.Title, "name", who)
	}
	state.Action = &kernel.CompactAction{
		ID:     "chores.done",
		Label:  kernel.NewTextf("Done: {title}", "title", first.Title),
		Method: http.MethodPost,
		Path:   "/api/chores/" + url.PathEscape(first.ID) + "/done",
	}
	return state, nil
}
