package controller

// SettingsInputDTO is the body of PUT /api/calendar/settings; an empty
// password keeps the stored one.
type SettingsInputDTO struct {
	CalendarURLs []string `json:"calendarUrls"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
}

// NewEventDTO is the body of POST /api/calendar/events. Timed events use
// RFC 3339 instants; all-day events use dates (YYYY-MM-DD) with an
// inclusive end date.
type NewEventDTO struct {
	CalendarURL string `json:"calendarUrl"`
	Summary     string `json:"summary"`
	Location    string `json:"location"`
	AllDay      bool   `json:"allDay"`
	Start       string `json:"start"`
	End         string `json:"end"`
}
