package todos

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/todos/adapter/controller"
	"trisse-home/internal/modules/todos/adapter/gateway"
	"trisse-home/internal/modules/todos/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the to-do rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the to-do module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	d := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids}
	return &Module{handlers: controller.NewHandlers(controller.UseCases{
		List: usecase.NewListTodos(d), Create: usecase.NewCreateTodo(d), Update: usecase.NewUpdateTodo(d),
		Delete: usecase.NewDeleteTodo(d), ClearCompleted: usecase.NewClearCompleted(d),
	})}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "todos" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"todos"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/todos", h.List)
	r.Protected("POST /api/todos", h.Create)
	r.Protected("POST /api/todos/clear-completed", h.ClearCompleted)
	r.Protected("PUT /api/todos/{id}", h.Update)
	r.Protected("DELETE /api/todos/{id}", h.Delete)
}
