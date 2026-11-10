package domain

import "trisse-home/internal/shared/kernel"

// Domain errors of the dashboard module.
var (
	ErrTabNotFound       = kernel.NewPublicError(kernel.ErrNotFound, "The tab does not exist.")
	ErrWidgetNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "The widget does not exist.")
	ErrLastTab           = kernel.NewPublicError(kernel.ErrConflict, "The last tab cannot be deleted.")
	ErrUnknownWidgetKind = kernel.NewPublicError(kernel.ErrInvalid, "This widget type is not available.")
	ErrTabOrderMismatch  = kernel.NewPublicError(kernel.ErrInvalid, "The new order must contain every tab exactly once.")
)
