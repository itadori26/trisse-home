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

// MailSettingsInputDTO is the body of PUT /api/parcels/mail (#166); an
// empty password keeps the stored one for the same server and user.
type MailSettingsInputDTO struct {
	Enabled         bool   `json:"enabled"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Folder          string `json:"folder"`
	IntervalMinutes int    `json:"intervalMinutes"`
}
