// Package presenter maps appearance responses to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import "trisse-home/internal/modules/appearance/usecase"

// AppearanceDTO is returned by GET/PUT /api/appearance and is also the body
// of PUT /api/appearance.
type AppearanceDTO struct {
	Theme           string `json:"theme" tstype:"'system' | 'light' | 'dark'"`
	AccentColor     string `json:"accentColor"`
	BackgroundImage string `json:"backgroundImage"`
	CustomCSS       string `json:"customCss"`
}

// Appearance maps an AppearanceResponse.
func Appearance(r usecase.AppearanceResponse) AppearanceDTO {
	return AppearanceDTO{Theme: r.Theme, AccentColor: r.AccentColor, BackgroundImage: r.BackgroundImage, CustomCSS: r.CustomCSS}
}
