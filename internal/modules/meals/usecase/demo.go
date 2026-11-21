package usecase

import (
	"context"
	"time"

	"waylight/internal/modules/meals/domain"
)

// SeedDemo plans the current week for the public demo, once.
func SeedDemo(ctx context.Context, d Deps) error {
	m := NewMeals(d)
	existing, err := m.Range(ctx, "", 7)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	day := func(n int) string { return now.AddDate(0, 0, n).Format(time.DateOnly) }
	for _, in := range []domain.Input{
		{Date: day(0), Meal: "dinner", Dish: "Spaghetti bolognese", Ingredients: []string{"Spaghetti", "Minced beef", "Chopped tomatoes", "Onions", "Parmesan"}},
		{Date: day(1), Meal: "dinner", Dish: "Vegetable curry with rice", Link: "https://www.example.org/recipes/vegetable-curry", Ingredients: []string{"Rice", "Coconut milk", "Curry paste", "Sweet potatoes", "Spinach"}},
		{Date: day(2), Meal: "lunch", Dish: "Pancakes", Ingredients: []string{"Flour", "Eggs", "Milk", "Apple sauce"}},
		{Date: day(2), Meal: "dinner", Dish: "Leftover curry"},
		{Date: day(3), Meal: "dinner", Dish: "Fish fingers with potatoes", Ingredients: []string{"Fish fingers", "Potatoes", "Peas"}},
		{Date: day(4), Meal: "dinner", Dish: "Pizza night", Ingredients: []string{"Pizza dough", "Mozzarella", "Passata", "Basil"}},
		{Date: day(5), Meal: "breakfast", Dish: "Family brunch", Ingredients: []string{"Bread rolls", "Eggs", "Orange juice"}},
	} {
		if _, err := m.Create(ctx, in, ""); err != nil {
			return err
		}
	}
	return nil
}
