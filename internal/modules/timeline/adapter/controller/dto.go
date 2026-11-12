package controller

// SettingsDTO is the body of PUT /api/timeline/settings.
type SettingsDTO struct {
	RetentionDays int `json:"retentionDays"`
}
