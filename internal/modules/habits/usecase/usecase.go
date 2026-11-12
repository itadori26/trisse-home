package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/habits/domain"
	"trisse-home/internal/shared/kernel"
)

// Habits is the input port of the module; every method is one use case (#106).
type Habits interface {
	// List returns the habits in order with their statistics.
	List(ctx context.Context) ([]HabitView, error)
	Create(ctx context.Context, in Input) (HabitView, error)
	Update(ctx context.Context, id string, in Input) (HabitView, error)
	Delete(ctx context.Context, id string) error
	// Toggle checks a habit off for a day (today or yesterday) or undoes it.
	Toggle(ctx context.Context, id string, date string) (HabitView, error)
}

// Input creates or changes a habit.
type Input struct {
	Name  string
	Emoji string
	Days  []int // weekdays 0–6, empty = every day
}

// HabitView is a habit with its statistics.
type HabitView struct {
	Habit domain.Habit
	Stats domain.Stats
}

// Repository persists habits and completions (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Habit, error)
	Get(ctx context.Context, id string) (domain.Habit, error)
	Save(ctx context.Context, h domain.Habit) error
	Delete(ctx context.Context, id string) error
	// Completions returns the completed days of the habits since a day.
	Completions(ctx context.Context, since domain.Day) (map[string]map[domain.Day]bool, error)
	SetCompletion(ctx context.Context, id string, day domain.Day, done bool) error
}

// Deps bundles the output ports. Location is the user's time zone (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewHabits returns the habit use cases.
func NewHabits(d Deps) Habits {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) today() domain.Day { return domain.DayOf(u.d.Clock.Now().In(u.d.Location)) }

// List implements Habits.
func (u *uc) List(ctx context.Context) ([]HabitView, error) {
	habits, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list habits: %w", err)
	}
	// The whole history is needed for the best streak; habits are small.
	done, err := u.d.Repo.Completions(ctx, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("list completions: %w", err)
	}
	today := u.today()
	out := make([]HabitView, 0, len(habits))
	for _, h := range habits {
		out = append(out, HabitView{Habit: h, Stats: domain.ComputeStats(h, done[h.ID], today)})
	}
	return out, nil
}

func (u *uc) view(ctx context.Context, h domain.Habit) (HabitView, error) {
	done, err := u.d.Repo.Completions(ctx, time.Time{})
	if err != nil {
		return HabitView{}, fmt.Errorf("list completions: %w", err)
	}
	return HabitView{Habit: h, Stats: domain.ComputeStats(h, done[h.ID], u.today())}, nil
}

// Create implements Habits.
func (u *uc) Create(ctx context.Context, in Input) (HabitView, error) {
	s, err := domain.NewSchedule(in.Days)
	if err != nil {
		return HabitView{}, err
	}
	all, err := u.d.Repo.List(ctx)
	if err != nil {
		return HabitView{}, fmt.Errorf("list habits: %w", err)
	}
	h, err := domain.New(u.d.IDs.NewID(), in.Name, in.Emoji, s, len(all), u.d.Clock.Now())
	if err != nil {
		return HabitView{}, err
	}
	if err := u.d.Repo.Save(ctx, h); err != nil {
		return HabitView{}, fmt.Errorf("save habit: %w", err)
	}
	return u.view(ctx, h)
}

// Update implements Habits.
func (u *uc) Update(ctx context.Context, id string, in Input) (HabitView, error) {
	h, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return HabitView{}, err
	}
	s, err := domain.NewSchedule(in.Days)
	if err != nil {
		return HabitView{}, err
	}
	if h, err = h.Update(in.Name, in.Emoji, s); err != nil {
		return HabitView{}, err
	}
	if err := u.d.Repo.Save(ctx, h); err != nil {
		return HabitView{}, fmt.Errorf("save habit: %w", err)
	}
	return u.view(ctx, h)
}

// Delete implements Habits.
func (u *uc) Delete(ctx context.Context, id string) error {
	if _, err := u.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	return u.d.Repo.Delete(ctx, id)
}

// Toggle implements Habits.
func (u *uc) Toggle(ctx context.Context, id string, date string) (HabitView, error) {
	h, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return HabitView{}, err
	}
	day, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return HabitView{}, domain.ErrDayNotEditable
	}
	if err := domain.EditableDay(day, u.today()); err != nil {
		return HabitView{}, err
	}
	done, err := u.d.Repo.Completions(ctx, day)
	if err != nil {
		return HabitView{}, fmt.Errorf("list completions: %w", err)
	}
	if err := u.d.Repo.SetCompletion(ctx, id, day, !done[id][day]); err != nil {
		return HabitView{}, fmt.Errorf("save completion: %w", err)
	}
	return u.view(ctx, h)
}
