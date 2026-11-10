package domain

import "trisse-home/internal/shared/kernel"

// Grid limits. The frontend uses the same 12-column grid.
const (
	GridColumns   = 12
	MaxWidgetRows = 24
	MaxGridRow    = 500
)

// Rect is the position and size of a widget in grid cells.
type Rect struct {
	X, Y, W, H int
}

// NewRect validates that the rectangle fits into the grid.
func NewRect(x, y, w, h int) (Rect, error) {
	switch {
	case w < 1 || w > GridColumns:
		return Rect{}, kernel.NewValidationError("w", "The width must be between 1 and 12 columns.")
	case h < 1 || h > MaxWidgetRows:
		return Rect{}, kernel.NewValidationError("h", "The height must be between 1 and 24 rows.")
	case x < 0 || x+w > GridColumns:
		return Rect{}, kernel.NewValidationError("x", "The widget must fit into the 12-column grid.")
	case y < 0 || y > MaxGridRow:
		return Rect{}, kernel.NewValidationError("y", "The row is out of range.")
	}
	return Rect{X: x, Y: y, W: w, H: h}, nil
}

// Overlaps reports whether two rectangles share at least one cell.
func (r Rect) Overlaps(o Rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}
