package controller

// SourceInputDTO is the body of POST/PUT /api/backup-status/sources.
type SourceInputDTO struct {
	Name        string  `json:"name"`
	MaxAgeHours float64 `json:"maxAgeHours"`
}
