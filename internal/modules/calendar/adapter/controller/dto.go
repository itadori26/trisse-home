package controller

// SettingsInputDTO is the body of PUT /api/calendar/settings; an empty
// password keeps the stored one.
type SettingsInputDTO struct {
	CalendarURLs []string `json:"calendarUrls"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
}
