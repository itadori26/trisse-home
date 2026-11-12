package controller

// RemoteSettingsInputDTO is the body of PUT /api/backups/remote/settings and
// POST /api/backups/remote/test; an empty password keeps the stored one.
type RemoteSettingsInputDTO struct {
	URL       string `json:"url"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Retention int    `json:"retention"`
}
