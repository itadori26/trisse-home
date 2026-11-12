package controller

// HabitInputDTO is the body of POST /api/habits and PUT /api/habits/{id}.
type HabitInputDTO struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	// Days are the weekdays 0 (Sunday) to 6; empty = every day.
	Days []int `json:"days"`
}

// ToggleDTO is the body of POST /api/habits/{id}/toggle.
type ToggleDTO struct {
	// Date is today or yesterday as YYYY-MM-DD.
	Date string `json:"date"`
}
