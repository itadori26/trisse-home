package domain

import (
	"regexp"

	"trisse-home/internal/shared/kernel"
)

// MaxConfigBytes limits the opaque widget configuration.
const MaxConfigBytes = 16 * 1024

var kindPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

// Widget is a tile on a tab. Its configuration is an opaque JSON document
// owned by the module that renders the widget.
type Widget struct {
	id     string
	tabID  string
	kind   string
	rect   Rect
	config string
}

// NewWidget creates a widget after validating kind and configuration.
func NewWidget(id, tabID, kind string, rect Rect, config string) (Widget, error) {
	if !kindPattern.MatchString(kind) {
		return Widget{}, kernel.NewValidationError("kind", "The widget type is invalid.")
	}
	w := Widget{id: id, tabID: tabID, kind: kind, rect: rect}
	if err := w.Configure(config); err != nil {
		return Widget{}, err
	}
	return w, nil
}

// RestoreWidget rebuilds a widget from persisted state.
func RestoreWidget(id, tabID, kind string, rect Rect, config string) Widget {
	return Widget{id: id, tabID: tabID, kind: kind, rect: rect, config: config}
}

// ID returns the identifier.
func (w Widget) ID() string { return w.id }

// TabID returns the tab the widget belongs to.
func (w Widget) TabID() string { return w.tabID }

// Kind returns the widget type, e.g. "weather".
func (w Widget) Kind() string { return w.kind }

// Rect returns position and size.
func (w Widget) Rect() Rect { return w.rect }

// Config returns the JSON configuration.
func (w Widget) Config() string { return w.config }

// Configure replaces the configuration. An empty configuration becomes "{}".
func (w *Widget) Configure(config string) error {
	if config == "" {
		config = "{}"
	}
	if len(config) > MaxConfigBytes {
		return kernel.NewValidationError("config", "The widget configuration is too large.")
	}
	w.config = config
	return nil
}

// Place moves and resizes the widget.
func (w *Widget) Place(r Rect) { w.rect = r }
