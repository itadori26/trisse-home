// Package controller contains the speedtest HTTP handlers.
//
// Ring: Interface Adapters.
package controller

// SettingsInputDTO is the body of PUT /api/speedtest/settings. An empty
// token keeps the stored one; thresholds of 0 disable notifications.
type SettingsInputDTO struct {
	URL         string  `json:"url"`
	Token       string  `json:"token"`
	MinDownMbps float64 `json:"minDownMbps"`
	MinUpMbps   float64 `json:"minUpMbps"`
}
