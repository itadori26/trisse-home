package domain

import (
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Template limits.
const (
	MaxTemplates       = 50
	MaxTemplateWidgets = 60
)

// Template errors.
var (
	ErrTemplateNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The template does not exist.")
	ErrTooManyTemplates = kernel.NewPublicError(kernel.ErrConflict, "At most 50 templates can be stored.")
)

// TemplateWidget is a widget inside a template (kind, place and config).
type TemplateWidget struct {
	Kind   string
	Rect   Rect
	Config string
}

// Template is a saved tab layout that new tabs can be created from.
type Template struct {
	ID      string
	Name    string
	Widgets []TemplateWidget
}

// NewTemplate captures the widgets of a tab under a name.
func NewTemplate(id, name string, widgets []Widget) (Template, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > MaxTabNameLength {
		return Template{}, kernel.NewValidationError("name", "The template name must be between 1 and 40 characters long.")
	}
	if len(widgets) > MaxTemplateWidgets {
		return Template{}, kernel.NewValidationError("tabId", "The tab has too many widgets for a template.")
	}
	t := Template{ID: id, Name: name, Widgets: make([]TemplateWidget, 0, len(widgets))}
	for _, w := range widgets {
		t.Widgets = append(t.Widgets, TemplateWidget{Kind: w.Kind(), Rect: w.Rect(), Config: w.Config()})
	}
	return t, nil
}
