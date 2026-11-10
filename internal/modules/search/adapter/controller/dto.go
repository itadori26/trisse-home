package controller

// DefaultProviderDTO is the body of PUT /api/search/settings.
type DefaultProviderDTO struct {
	DefaultProvider string `json:"defaultProvider"`
}
