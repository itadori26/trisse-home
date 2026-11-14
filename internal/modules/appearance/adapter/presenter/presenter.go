package presenter

import (
	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/modules/appearance/usecase"
)

// AppearanceDTO is returned by GET/PUT /api/appearance and is also the body
// of PUT /api/appearance.
type AppearanceDTO struct {
	Theme           string `json:"theme" tstype:"'system' | 'light' | 'dark'"`
	AccentColor     string `json:"accentColor"`
	BackgroundImage string `json:"backgroundImage"`
	CustomCSS       string `json:"customCss"`
	Language        string `json:"language" tstype:"'' | 'en' | 'de'"`
	// GreetingName is the name in the greeting (#189); may be empty.
	GreetingName string `json:"greetingName"`
	// Launcher is what the launcher lists (#189).
	Launcher string `json:"launcher" tstype:"'both' | 'services' | 'bookmarks'"`
}

// Appearance maps an AppearanceResponse.
func Appearance(r usecase.AppearanceResponse) AppearanceDTO {
	return AppearanceDTO{Theme: r.Theme, AccentColor: r.AccentColor, BackgroundImage: r.BackgroundImage, CustomCSS: r.CustomCSS, Language: r.Language,
		GreetingName: r.GreetingName, Launcher: r.Launcher}
}

// ThemeDTO is a gallery theme.
type ThemeDTO struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Author  string            `json:"author"`
	Scheme  string            `json:"scheme" tstype:"'light' | 'dark'"`
	Tokens  map[string]string `json:"tokens"`
	BuiltIn bool              `json:"builtIn"`
}

// GalleryDTO is the theme gallery (#111).
type GalleryDTO struct {
	Themes []ThemeDTO `json:"themes"`
	// Applied is the id of the theme in use ("" = own or no custom CSS).
	Applied    string `json:"applied"`
	CanRestore bool   `json:"canRestore"`
}

// Gallery maps the gallery.
func Gallery(g usecase.Gallery) GalleryDTO {
	out := GalleryDTO{Themes: make([]ThemeDTO, 0, len(g.Themes)), Applied: g.Applied, CanRestore: g.CanRestore}
	for _, t := range g.Themes {
		out.Themes = append(out.Themes, Theme(t))
	}
	return out
}

// Theme maps a theme.
func Theme(t domain.ThemeDef) ThemeDTO {
	return ThemeDTO{ID: t.ID, Name: t.Name, Author: t.Author, Scheme: string(t.Scheme), Tokens: t.Tokens, BuiltIn: t.BuiltIn}
}
