package controller

// SettingsInputDTO is the body of PUT /api/adguard/settings?provider=….
// An empty password keeps the stored one.
type SettingsInputDTO struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}
