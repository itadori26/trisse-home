package presenter

import "trisse-home/internal/modules/backup/usecase"

// RemoteSettingsDTO is the off-site target; the password is never returned.
type RemoteSettingsDTO struct {
	Configured  bool   `json:"configured"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	HasPassword bool   `json:"hasPassword"`
	Retention   int    `json:"retention"`
}

// RemoteSettings maps the settings view.
func RemoteSettings(v usecase.RemoteSettingsView) RemoteSettingsDTO { return RemoteSettingsDTO(v) }
