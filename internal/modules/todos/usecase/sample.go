package usecase

import (
	"context"
	"fmt"
	"time"
)

// SeedSampleData creates example to-dos if there are none (input port).
type SeedSampleData interface {
	Execute(ctx context.Context) error
}

type seedSampleData struct{ d Deps }

// NewSeedSampleData returns the SeedSampleData use case.
func NewSeedSampleData(d Deps) SeedSampleData { return &seedSampleData{d: d} }

// Execute implements SeedSampleData. Due dates are relative to today so the
// list always looks current.
func (uc *seedSampleData) Execute(ctx context.Context) error {
	existing, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list todos: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}
	today := uc.d.Clock.Now()
	samples := []struct {
		title string
		due   time.Time
		done  bool
	}{
		{"Set up trisse-home", time.Time{}, true},
		{"Try edit mode: drag and resize a widget", today, false},
		{"Add your own services under Settings → Services", today.AddDate(0, 0, 1), false},
		{"Connect your calendar under Settings → Calendar", today.AddDate(0, 0, 3), false},
		{"Choose your weather location (⚙ on the weather widget)", time.Time{}, false},
	}
	create, update := NewCreateTodo(uc.d), NewUpdateTodo(uc.d)
	for _, s := range samples {
		t, err := create.Execute(ctx, TodoInput{Title: s.title, Due: s.due})
		if err != nil {
			return err
		}
		if s.done {
			if _, err := update.Execute(ctx, t.ID, TodoInput{Title: s.title}, true); err != nil {
				return err
			}
		}
	}
	return nil
}
