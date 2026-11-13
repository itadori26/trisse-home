package controller

// CountdownInputDTO is the body of POST /api/countdowns and
// PUT /api/countdowns/{id}.
type CountdownInputDTO struct {
	Title string `json:"title"`
	Emoji string `json:"emoji"`
	// Date is YYYY-MM-DD.
	Date string `json:"date"`
	// Time is HH:MM, or empty for the whole day.
	Time string `json:"time"`
}
