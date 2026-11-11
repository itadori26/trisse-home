package controller

// SettingsInputDTO is the body of PUT /api/shopping/settings; an empty
// password keeps the stored one.
type SettingsInputDTO struct {
	ListURL  string `json:"listUrl"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewItemDTO is the body of POST /api/shopping/items.
type NewItemDTO struct {
	Summary string `json:"summary"`
}

// ItemStateDTO is the body of PUT /api/shopping/items/{id}.
type ItemStateDTO struct {
	Done bool `json:"done"`
}
