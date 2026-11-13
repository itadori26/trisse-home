package controller

// SettingsInputDTO is the body of PUT /api/proxmox/settings. An empty
// secret keeps the stored one.
type SettingsInputDTO struct {
	URL         string `json:"url"`
	TokenID     string `json:"tokenId"`
	Secret      string `json:"secret"`
	Fingerprint string `json:"fingerprint"`
}
