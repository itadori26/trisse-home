package controller

// NoteInputDTO is the body of POST /api/notes and PUT /api/notes/{id}.
type NoteInputDTO struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Pinned bool   `json:"pinned"`
}
