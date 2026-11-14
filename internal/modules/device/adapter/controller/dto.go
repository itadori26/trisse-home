package controller

// ModeDTO is the body of PUT /api/device/display-mode.
type ModeDTO struct {
	Mode string `json:"mode" tstype:"'headless' | 'kiosk'"`
}
