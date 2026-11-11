package controller

// SettingsInputDTO is the body of PUT /api/parcels/settings; an empty key
// keeps the stored one.
type SettingsInputDTO struct {
	DHLAPIKey         string `json:"dhlApiKey"`
	HideDeliveredDays int    `json:"hideDeliveredDays"`
}

// NewParcelDTO is the body of POST /api/parcels.
type NewParcelDTO struct {
	TrackingNumber string `json:"trackingNumber"`
	Carrier        string `json:"carrier"`
	Label          string `json:"label"`
}
