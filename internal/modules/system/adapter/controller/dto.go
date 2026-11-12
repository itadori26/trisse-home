package controller

// UpdateCheckDTO is the body of PUT /api/system/update-check.
type UpdateCheckDTO struct {
	Enabled bool `json:"enabled"`
}

// ForecastSettingsDTO is the body of PUT /api/system/forecast-settings.
type ForecastSettingsDTO struct {
	WindowDays int `json:"windowDays"`
	AlertDays  int `json:"alertDays"`
}
