package controller

// EntryInputDTO is the body of POST /api/meals and PUT /api/meals/{id}.
type EntryInputDTO struct {
	// Date is YYYY-MM-DD.
	Date string `json:"date"`
	// Meal is "breakfast", "lunch", "dinner" or empty.
	Meal        string   `json:"meal"`
	Dish        string   `json:"dish"`
	Link        string   `json:"link"`
	Ingredients []string `json:"ingredients"`
	// Visibility is "private", "everyone" or "group:<id>". Empty shares a
	// new dish with everyone and keeps the visibility of an existing one.
	Visibility string `json:"visibility,omitempty"`
}
