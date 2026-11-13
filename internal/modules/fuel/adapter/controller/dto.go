package controller

// SettingsInputDTO is the body of PUT /api/fuel/settings; an empty API key
// keeps the stored one.
type SettingsInputDTO struct {
	APIKey   string  `json:"apiKey"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Place    string  `json:"place"`
	RadiusKm float64 `json:"radiusKm"`
	Fuel     string  `json:"fuel" tstype:"'e5' | 'e10' | 'diesel'"`
}

// FavouriteDTO is a favourite station.
type FavouriteDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
