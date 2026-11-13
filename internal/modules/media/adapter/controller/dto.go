package controller

// SettingsInputDTO is the body of PUT /api/media/settings; an empty token
// keeps the stored one (for the same provider).
type SettingsInputDTO struct {
	Provider string `json:"provider" tstype:"'jellyfin' | 'plex'"`
	URL      string `json:"url"`
	Token    string `json:"token"`
}
