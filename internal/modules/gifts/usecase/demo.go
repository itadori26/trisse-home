package usecase

import (
	"context"

	"waylight/internal/modules/gifts/domain"
)

// SeedDemo adds ideas for the people of the demo birthdays, once.
func SeedDemo(ctx context.Context, d Deps) error {
	g := NewGifts(d)
	existing, err := g.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	for _, in := range []domain.Input{
		{Person: "Grandma Hilde", Title: "Large-print crossword book"},
		{Person: "Grandma Hilde", Title: "Photo calendar with the kids", Note: "Pick photos from the summer holiday"},
		{Person: "Anna Schröder", Title: "Pottery course voucher", Link: "https://www.example.org/pottery"},
		{Person: "Paul", Title: "Wooden marble run"},
		{Person: "Paul", Title: "Picture book about dinosaurs", Bought: true},
	} {
		if _, err := g.Create(ctx, in); err != nil {
			return err
		}
	}
	return nil
}
