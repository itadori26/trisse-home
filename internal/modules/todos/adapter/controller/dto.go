package controller

// TodoInputDTO is the body of POST /api/todos and PUT /api/todos/{id}.
// Due is a date (YYYY-MM-DD) or empty.
type TodoInputDTO struct {
	Title string `json:"title"`
	Due   string `json:"due"`
	Done  bool   `json:"done"`
}

// ClearedDTO is returned by POST /api/todos/clear-completed.
type ClearedDTO struct {
	Removed int `json:"removed"`
}
