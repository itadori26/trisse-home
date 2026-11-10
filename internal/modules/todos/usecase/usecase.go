package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/shared/kernel"
)

// TodoInput carries the editable fields of a to-do.
type TodoInput struct {
	Title string
	Due   time.Time
}

// TodoView is a to-do at the use case boundary.
type TodoView struct {
	ID          string
	Title       string
	Due         time.Time
	Done        bool
	Overdue     bool
	Position    int
	CreatedAt   time.Time
	CompletedAt time.Time
}

// ListTodos returns open to-dos first (by position), then done ones (input port).
type ListTodos interface {
	Execute(ctx context.Context, includeDone bool) ([]TodoView, error)
}

// CreateTodo appends a to-do (input port).
type CreateTodo interface {
	Execute(ctx context.Context, in TodoInput) (TodoView, error)
}

// UpdateTodo edits text, due date and completion (input port).
type UpdateTodo interface {
	Execute(ctx context.Context, id string, in TodoInput, done bool) (TodoView, error)
}

// DeleteTodo removes a to-do (input port).
type DeleteTodo interface {
	Execute(ctx context.Context, id string) error
}

// ClearCompleted removes all done to-dos (input port).
type ClearCompleted interface {
	Execute(ctx context.Context) (int, error)
}

// Repository persists to-dos (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Todo, error)
	Get(ctx context.Context, id string) (domain.Todo, error)
	Save(ctx context.Context, t domain.Todo) error
	Delete(ctx context.Context, ids ...string) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo  Repository
	Clock kernel.Clock
	IDs   kernel.IDGenerator
}

func (d Deps) view(t domain.Todo) TodoView {
	return TodoView{
		ID: t.ID(), Title: t.Title(), Due: t.Due(), Done: t.Done(), Overdue: t.IsOverdue(d.Clock.Now()),
		Position: t.Position(), CreatedAt: t.CreatedAt(), CompletedAt: t.CompletedAt(),
	}
}

type listTodos struct{ d Deps }

// NewListTodos returns the ListTodos use case.
func NewListTodos(d Deps) ListTodos { return &listTodos{d: d} }

// Execute implements ListTodos.
func (uc *listTodos) Execute(ctx context.Context, includeDone bool) ([]TodoView, error) {
	todos, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	slices.SortStableFunc(todos, func(a, b domain.Todo) int {
		switch {
		case a.Done() != b.Done():
			if a.Done() {
				return 1
			}
			return -1
		case a.Done():
			return b.CompletedAt().Compare(a.CompletedAt())
		default:
			return a.Position() - b.Position()
		}
	})
	out := make([]TodoView, 0, len(todos))
	for _, t := range todos {
		if includeDone || !t.Done() {
			out = append(out, uc.d.view(t))
		}
	}
	return out, nil
}

type createTodo struct{ d Deps }

// NewCreateTodo returns the CreateTodo use case.
func NewCreateTodo(d Deps) CreateTodo { return &createTodo{d: d} }

// Execute implements CreateTodo.
func (uc *createTodo) Execute(ctx context.Context, in TodoInput) (TodoView, error) {
	todos, err := uc.d.Repo.List(ctx)
	if err != nil {
		return TodoView{}, fmt.Errorf("list todos: %w", err)
	}
	next := 0
	for _, t := range todos {
		next = max(next, t.Position()+1)
	}
	t, err := domain.NewTodo(uc.d.IDs.NewID(), in.Title, in.Due, next, uc.d.Clock.Now())
	if err != nil {
		return TodoView{}, err
	}
	if err := uc.d.Repo.Save(ctx, t); err != nil {
		return TodoView{}, fmt.Errorf("save todo: %w", err)
	}
	return uc.d.view(t), nil
}

type updateTodo struct{ d Deps }

// NewUpdateTodo returns the UpdateTodo use case.
func NewUpdateTodo(d Deps) UpdateTodo { return &updateTodo{d: d} }

// Execute implements UpdateTodo.
func (uc *updateTodo) Execute(ctx context.Context, id string, in TodoInput, done bool) (TodoView, error) {
	t, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return TodoView{}, err
	}
	if err := t.Edit(in.Title, in.Due); err != nil {
		return TodoView{}, err
	}
	t.SetDone(done, uc.d.Clock.Now())
	if err := uc.d.Repo.Save(ctx, t); err != nil {
		return TodoView{}, fmt.Errorf("save todo: %w", err)
	}
	return uc.d.view(t), nil
}

type deleteTodo struct{ d Deps }

// NewDeleteTodo returns the DeleteTodo use case.
func NewDeleteTodo(d Deps) DeleteTodo { return &deleteTodo{d: d} }

// Execute implements DeleteTodo.
func (uc *deleteTodo) Execute(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	if err := uc.d.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	return nil
}

type clearCompleted struct{ d Deps }

// NewClearCompleted returns the ClearCompleted use case.
func NewClearCompleted(d Deps) ClearCompleted { return &clearCompleted{d: d} }

// Execute implements ClearCompleted.
func (uc *clearCompleted) Execute(ctx context.Context) (int, error) {
	todos, err := uc.d.Repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list todos: %w", err)
	}
	var ids []string
	for _, t := range todos {
		if t.Done() {
			ids = append(ids, t.ID())
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	if err := uc.d.Repo.Delete(ctx, ids...); err != nil {
		return 0, fmt.Errorf("delete todos: %w", err)
	}
	return len(ids), nil
}
