package presenter

import "trisse-home/internal/modules/appearance/usecase"

// AppearanceDTO is returned by GET/PUT /api/appearance and is also the body
// of PUT /api/appearance.
type AppearanceDTO struct {
	Theme           string `json:"theme" tstype:"'system' | 'light' | 'dark'"`
	AccentColor     string `json:"accentColor"`
	BackgroundImage string `json:"backgroundImage"`
	CustomCSS       string `json:"customCss"`
	Language        string `json:"language" tstype:"'' | 'en' | 'de'"`
	// Greeting shows the greeting line below the top bar (#167).
	Greeting     bool   `json:"greeting"`
	GreetingName string `json:"greetingName"`
}

// Appearance maps an AppearanceResponse.
func Appearance(r usecase.AppearanceResponse) AppearanceDTO {
	return AppearanceDTO{Theme: r.Theme, AccentColor: r.AccentColor, BackgroundImage: r.BackgroundImage, CustomCSS: r.CustomCSS, Language: r.Language,
		Greeting: r.Greeting, GreetingName: r.GreetingName}
}
