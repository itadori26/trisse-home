package controller

// SettingsInputDTO is the body of PUT /api/gamedeals/settings.
type SettingsInputDTO struct {
	// SteamID is the 64-bit ID of a profile with a public wishlist; empty disables Steam.
	SteamID     string `json:"steamId"`
	Country     string `json:"country"`
	MinDiscount int    `json:"minDiscount"`
	Epic        bool   `json:"epic"`
}
