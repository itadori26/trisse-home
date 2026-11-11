package controller

// UpdateCheckDTO is the body of PUT /api/system/update-check.
type UpdateCheckDTO struct {
	Enabled bool `json:"enabled"`
}
