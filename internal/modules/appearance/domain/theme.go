package domain

import (
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Theme gallery (#111): a theme sets design tokens only, never layout or
// free CSS, so app updates cannot break it and an imported theme cannot load
// anything or hide parts of the page.

// ThemeTokens are the design tokens a theme may set (see web tokens.css).
var ThemeTokens = []string{
	"--color-bg", "--color-surface", "--color-surface-2", "--color-border", "--color-text", "--color-text-muted",
	"--color-accent", "--color-accent-contrast", "--color-danger", "--color-success", "--color-warning",
	"--radius", "--radius-small",
}

// requiredTokens make a theme readable.
var requiredTokens = []string{"--color-bg", "--color-surface", "--color-text", "--color-accent"}

var (
	themeID    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,39}$`)
	colorValue = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	radiusVal  = regexp.MustCompile(`^(?:[0-9]|[1-2][0-9]|3[0-2])px$`)
)

// ThemeCSSMarker starts the custom CSS written by applying a theme.
const ThemeCSSMarker = "/* trisse-home theme: "

// ThemeDef is a gallery theme.
type ThemeDef struct {
	ID     string
	Name   string
	Author string
	// Scheme is light or dark: the base the tokens are meant for.
	Scheme Theme
	Tokens map[string]string
	// BuiltIn themes ship with the app and cannot be deleted.
	BuiltIn bool
}

// ErrThemeNotFound is returned for unknown themes.
var ErrThemeNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The theme does not exist.")

// NewThemeDef validates a theme.
func NewThemeDef(id, name, author string, scheme Theme, tokens map[string]string) (ThemeDef, error) {
	t := ThemeDef{ID: strings.ToLower(strings.TrimSpace(id)), Name: strings.TrimSpace(name), Author: strings.TrimSpace(author), Scheme: scheme, Tokens: map[string]string{}}
	switch {
	case !themeID.MatchString(t.ID):
		return ThemeDef{}, kernel.NewValidationError("id", "The theme id must be 2–40 lowercase letters, digits or '-'.")
	case t.Name == "" || utf8.RuneCountInString(t.Name) > 40:
		return ThemeDef{}, kernel.NewValidationError("name", "The theme needs a name of at most 40 characters.")
	case utf8.RuneCountInString(t.Author) > 60:
		return ThemeDef{}, kernel.NewValidationError("author", "The author must be at most 60 characters long.")
	case scheme != ThemeLight && scheme != ThemeDark:
		return ThemeDef{}, kernel.NewValidationError("scheme", "The scheme must be light or dark.")
	}
	for k, v := range tokens {
		v = strings.TrimSpace(v)
		switch {
		case !slices.Contains(ThemeTokens, k):
			return ThemeDef{}, kernel.NewValidationErrorf("tokens", "Unknown token {token}.", "token", k)
		case strings.HasPrefix(k, "--radius") && !radiusVal.MatchString(v),
			!strings.HasPrefix(k, "--radius") && !colorValue.MatchString(v):
			return ThemeDef{}, kernel.NewValidationErrorf("tokens", "Invalid value for {token}.", "token", k)
		}
		t.Tokens[k] = strings.ToLower(v)
	}
	for _, k := range requiredTokens {
		if _, ok := t.Tokens[k]; !ok {
			return ThemeDef{}, kernel.NewValidationErrorf("tokens", "The theme must set {token}.", "token", k)
		}
	}
	return t, nil
}

// CSS is the custom CSS that applies the theme: the tokens on the root for
// both schemes, marked so a later apply knows it is not the user's own CSS.
func (t ThemeDef) CSS() string {
	keys := make([]string, 0, len(t.Tokens))
	for k := range t.Tokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(ThemeCSSMarker + t.ID + " */\n:root:root,\n:root:root[data-theme] {\n")
	for _, k := range keys {
		b.WriteString("  " + k + ": " + t.Tokens[k] + ";\n")
	}
	b.WriteString("}\n")
	return b.String()
}

// IsThemeCSS reports whether custom CSS was written by applying a theme.
func IsThemeCSS(css string) bool { return strings.HasPrefix(css, ThemeCSSMarker) }

// AppliedTheme returns the id of the theme the custom CSS comes from, or "".
func AppliedTheme(css string) string {
	if !IsThemeCSS(css) {
		return ""
	}
	rest := strings.TrimPrefix(css, ThemeCSSMarker)
	id, _, _ := strings.Cut(rest, " ")
	return id
}

// WithStyle returns the appearance with other custom CSS and accent color
// and, unless scheme is empty, another scheme. A theme brings its own accent,
// so applying one clears the accent color (it would override the theme).
func (a Appearance) WithStyle(css string, scheme Theme, accentColor string) (Appearance, error) {
	if scheme == "" {
		scheme = a.theme
	}
	next, err := NewAppearance(scheme, accentColor, a.backgroundImage, css)
	if err != nil {
		return Appearance{}, err
	}
	next.language, next.greetingName, next.launcher = a.language, a.greetingName, a.launcher
	return next, nil
}
