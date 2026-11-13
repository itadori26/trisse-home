package controller

// DefaultProviderDTO is the body of PUT /api/search/settings.
type DefaultProviderDTO struct {
	DefaultProvider string `json:"defaultProvider"`
}

// BarOptionsDTO is the body of PUT /api/search/bar (#221).
type BarOptionsDTO struct {
	OpenInNewTab    bool `json:"openInNewTab"`
	CapturePrefixes struct {
		Todo     string `json:"todo"`
		Note     string `json:"note"`
		Bookmark string `json:"bookmark"`
	} `json:"capturePrefixes"`
}
