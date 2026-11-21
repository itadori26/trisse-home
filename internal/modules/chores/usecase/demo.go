package usecase

import (
	"context"
	"time"

	"waylight/internal/modules/chores/domain"
)

// SeedDemo adds household chores with a rotation to the public demo, once.
func SeedDemo(ctx context.Context, d Deps) error {
	c := NewChores(d)
	existing, err := c.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	day := func(n int) string { return now.AddDate(0, 0, n).Format(time.DateOnly) }
	family := []string{"Anna", "Ben", "Mia"}
	for _, in := range []domain.Input{
		{Title: "Empty the dishwasher", EveryDays: 1, Due: day(0), Assignees: family},
		{Title: "Water the plants", EveryDays: 3, Due: day(0), Assignees: []string{"Mia"}},
		{Title: "Take out the recycling", EveryDays: 7, Due: day(-1), Assignees: []string{"Ben", "Anna"}},
		{Title: "Vacuum the living room", EveryDays: 7, Due: day(2), Assignees: family},
		{Title: "Change the bed sheets", EveryDays: 14, Due: day(5)},
	} {
		if _, err := c.Create(ctx, in, ""); err != nil {
			return err
		}
	}
	return nil
}
