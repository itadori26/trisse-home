// Package domain holds the Appearance entity and its validation rules.
//
// Ring: Entities.
package domain

import (
	"net/url"
	"regexp"
	"strings"

	"trisse-home/internal/shared/kernel"
)

// Limits of user-provided values.
const (
	MaxCustomCSSBytes = 64 * 1024
	MaxURLLength      = 2048
)

// Theme is the color scheme preference.
type Theme string

// Supported themes.
const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Appearance holds the visual preferences. The zero value is not valid; use
// Default or NewAppearance.
type Appearance struct {
	theme           Theme
	accentColor     string
	backgroundImage string
	customCSS       string
}

// Default returns the appearance of a fresh installation.
func Default() Appearance { return Appearance{theme: ThemeSystem} }

// NewAppearance validates and creates an Appearance. Empty accentColor and
// backgroundImage mean "use the built-in default".
func NewAppearance(theme Theme, accentColor, backgroundImage, customCSS string) (Appearance, error) {
	switch theme {
	case ThemeSystem, ThemeLight, ThemeDark:
	default:
		return Appearance{}, kernel.NewValidationError("theme", "The theme must be system, light or dark.")
	}
	accentColor = strings.TrimSpace(accentColor)
	if accentColor != "" && !hexColor.MatchString(accentColor) {
		return Appearance{}, kernel.NewValidationError("accentColor", "The accent color must look like #3b6ef5.")
	}
	backgroundImage = strings.TrimSpace(backgroundImage)
	if backgroundImage != "" && !isHTTPURL(backgroundImage) {
		return Appearance{}, kernel.NewValidationError("backgroundImage", "The background image must be an http(s) URL.")
	}
	if len(customCSS) > MaxCustomCSSBytes {
		return Appearance{}, kernel.NewValidationError("customCss", "The custom CSS must be at most 64 KiB.")
	}
	return Appearance{theme: theme, accentColor: strings.ToLower(accentColor), backgroundImage: backgroundImage, customCSS: customCSS}, nil
}

func isHTTPURL(raw string) bool {
	if len(raw) > MaxURLLength || strings.ContainsAny(raw, "\"'()\\ \n") {
		// Quotes, parentheses and whitespace could break out of url(...) in CSS.
		return false
	}
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// Theme returns the color scheme.
func (a Appearance) Theme() Theme { return a.theme }

// AccentColor returns the accent color (#rrggbb) or "".
func (a Appearance) AccentColor() string { return a.accentColor }

// BackgroundImage returns the background image URL or "".
func (a Appearance) BackgroundImage() string { return a.backgroundImage }

// CustomCSS returns the user's CSS.
func (a Appearance) CustomCSS() string { return a.customCSS }
