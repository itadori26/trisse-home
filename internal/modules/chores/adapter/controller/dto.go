package controller

// ChoreInputDTO is the body of POST /api/chores and PUT /api/chores/{id}.
type ChoreInputDTO struct {
	Title     string `json:"title"`
	EveryDays int    `json:"everyDays"`
	// Due is the next due day, YYYY-MM-DD.
	Due string `json:"due"`
	// Assignees take turns in this order, starting with the first.
	Assignees []string `json:"assignees"`
	// Visibility is "private", "everyone" or "group:<id>". Empty shares a
	// new chore with everyone and keeps the visibility of an existing one.
	Visibility string `json:"visibility,omitempty"`
}
