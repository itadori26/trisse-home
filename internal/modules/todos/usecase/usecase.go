package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/shared/kernel"
)

// TodoInput carries the editable fields of a to-do. Nil Repeat or
// Reminder keep the stored values (older clients only send title and due).
type TodoInput struct {
	Title    string
	Due      time.Time
	Repeat   *RepeatInput
	Reminder *ReminderInput
}

// RepeatInput describes a recurrence (#68); an empty Kind means none.
type RepeatInput struct {
	Kind  string
	Every int
	Days  []int
}

// ReminderInput describes a reminder (#68); an empty At means none.
type ReminderInput struct {
	At         string
	DaysBefore int
}

// RepeatView is a recurrence at the use case boundary.
type RepeatView struct {
	Kind  string
	Every int
	Days  []int
}

// ReminderView is a reminder at the use case boundary.
type ReminderView struct {
	At         string
	DaysBefore int
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
	Repeat      RepeatView
	Reminder    ReminderView
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

// SendDueReminders pushes reminders of open to-dos that are due (input
// port, run every minute); it returns how many were sent.
type SendDueReminders interface {
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
	// Location is the user's time zone for completion days and reminder
	// times (nil = time.Local).
	Location *time.Location
	// Alerts delivers reminders (nil = reminders are not sent).
	Alerts kernel.AlertSink
}

func (d Deps) loc() *time.Location {
	if d.Location == nil {
		return time.Local
	}
	return d.Location
}

// applySchedule sets recurrence and reminder from the input; nil fields
// keep the current values.
func applySchedule(t *domain.Todo, in TodoInput) error {
	repeat, reminder := t.Repeat(), t.Reminder()
	var err error
	if in.Repeat != nil {
		if repeat, err = domain.NewRecurrence(domain.RepeatKind(in.Repeat.Kind), in.Repeat.Every, in.Repeat.Days); err != nil {
			return err
		}
	}
	if in.Reminder != nil {
		if reminder, err = domain.NewReminder(in.Reminder.At, in.Reminder.DaysBefore); err != nil {
			return err
		}
	}
	return t.Schedule(repeat, reminder)
}

func nextPosition(todos []domain.Todo) int {
	next := 0
	for _, t := range todos {
		next = max(next, t.Position()+1)
	}
	return next
}

func (d Deps) view(t domain.Todo) TodoView {
	return TodoView{
		ID: t.ID(), Title: t.Title(), Due: t.Due(), Done: t.Done(), Overdue: t.IsOverdue(d.Clock.Now()),
		Position: t.Position(), CreatedAt: t.CreatedAt(), CompletedAt: t.CompletedAt(),
		Repeat:   RepeatView{Kind: string(t.Repeat().Kind()), Every: t.Repeat().Every(), Days: t.Repeat().Days()},
		Reminder: ReminderView{At: t.Reminder().At(), DaysBefore: t.Reminder().DaysBefore()},
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
	t, err := domain.NewTodo(uc.d.IDs.NewID(), in.Title, in.Due, nextPosition(todos), uc.d.Clock.Now())
	if err != nil {
		return TodoView{}, err
	}
	if err := applySchedule(&t, in); err != nil {
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
	if err := applySchedule(&t, in); err != nil {
		return TodoView{}, err
	}
	now := uc.d.Clock.Now()
	completing := done && !t.Done()
	t.SetDone(done, now)
	if completing {
		// Completing a recurring to-do creates its next occurrence (#68).
		if err := uc.createNext(ctx, t, now); err != nil {
			return TodoView{}, err
		}
	}
	if err := uc.d.Repo.Save(ctx, t); err != nil {
		return TodoView{}, fmt.Errorf("save todo: %w", err)
	}
	return uc.d.view(t), nil
}

func (uc *updateTodo) createNext(ctx context.Context, t domain.Todo, now time.Time) error {
	if !t.Repeat().Repeats() {
		return nil
	}
	todos, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list todos: %w", err)
	}
	next, _ := t.NextOccurrence(uc.d.IDs.NewID(), nextPosition(todos), now.In(uc.d.loc()), now)
	if err := uc.d.Repo.Save(ctx, next); err != nil {
		return fmt.Errorf("save next occurrence: %w", err)
	}
	return nil
}

type sendDueReminders struct{ d Deps }

// NewSendDueReminders returns the SendDueReminders use case.
func NewSendDueReminders(d Deps) SendDueReminders { return &sendDueReminders{d: d} }

// Execute implements SendDueReminders.
func (uc *sendDueReminders) Execute(ctx context.Context) (int, error) {
	if uc.d.Alerts == nil {
		return 0, nil
	}
	todos, err := uc.d.Repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("list todos: %w", err)
	}
	now, loc := uc.d.Clock.Now(), uc.d.loc()
	sent := 0
	for _, t := range todos {
		if !t.ReminderDue(now, loc) {
			continue
		}
		// Mark first: a failed save must not lead to a push every minute.
		t.MarkReminded()
		if err := uc.d.Repo.Save(ctx, t); err != nil {
			return sent, fmt.Errorf("save todo: %w", err)
		}
		uc.d.Alerts.Alert(ctx, kernel.Alert{Kind: "reminder", Title: t.Title(), Body: reminderBody(t, now.In(loc))})
		sent++
	}
	return sent, nil
}

// reminderBody describes the due date relative to today (English; push
// texts follow the server language like other alerts).
func reminderBody(t domain.Todo, today time.Time) string {
	y, m, d := today.Date()
	days := int(t.Due().Sub(time.Date(y, m, d, 0, 0, 0, 0, time.UTC)).Hours() / 24)
	switch {
	case days == 0:
		return "To-do due today."
	case days == 1:
		return "To-do due tomorrow."
	case days < 0:
		return fmt.Sprintf("To-do overdue since %s.", t.Due().Format("2 Jan 2006"))
	default:
		return fmt.Sprintf("To-do due on %s (in %d days).", t.Due().Format("Mon, 2 Jan"), days)
	}
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
