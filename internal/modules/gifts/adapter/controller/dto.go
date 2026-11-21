package controller

// IdeaInputDTO is the body of POST /api/gifts and PUT /api/gifts/{id}.
type IdeaInputDTO struct {
	Person string `json:"person"`
	Title  string `json:"title"`
	Link   string `json:"link"`
	Note   string `json:"note"`
	Bought bool   `json:"bought"`
}
