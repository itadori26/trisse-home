package controller

// SettingsInputDTO is the body of PUT /api/immich/settings; an empty API
// key keeps the stored one.
type SettingsInputDTO struct {
	URL    string `json:"url"`
	APIKey string `json:"apiKey"`
}
