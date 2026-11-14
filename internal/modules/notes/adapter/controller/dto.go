package controller

// NoteInputDTO is the body of POST /api/notes and PUT /api/notes/{id}.
type NoteInputDTO struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Pinned bool   `json:"pinned"`
	// ID (create) is a client-generated ID; replaying the create returns the
	// existing note (#80).
	ID string `json:"id,omitempty"`
	// Version (update) is the version the client edited; a different stored
	// version answers 409 Conflict. 0/omitted = no check.
	Version int `json:"version,omitempty"`
}
