package controller

// SettingsInputDTO is the body of PUT /api/adguard/settings?provider=….
// An empty password keeps the stored one.
type SettingsInputDTO struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// PauseDTO is the body of POST /api/adguard/pause?provider=… (#177).
type PauseDTO struct {
	// Minutes is 5, 30 or 60.
	Minutes int `json:"minutes"`
}
