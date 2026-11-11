package domain

import "trisse-home/internal/shared/kernel"

// Domain errors of the auth module.
var (
	ErrAccountNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "No account has been set up yet.")
	ErrAlreadySetUp       = kernel.NewPublicError(kernel.ErrConflict, "The account has already been set up.")
	ErrSetupRequired      = kernel.NewPublicError(kernel.ErrForbidden, "The account must be set up first.")
	ErrInvalidSetupToken  = kernel.NewPublicError(kernel.ErrUnauthorized, "The setup token is invalid.")
	ErrInvalidCredentials = kernel.NewPublicError(kernel.ErrUnauthorized, "The password is incorrect.")
	ErrSessionInvalid     = kernel.NewPublicError(kernel.ErrUnauthorized, "Your session has expired. Please log in again.")
	ErrSessionNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "Session not found.")
	ErrSessionNotListed   = kernel.NewPublicError(kernel.ErrNotFound, "The session does not exist (anymore).")
	ErrTooManyAttempts    = kernel.NewPublicError(kernel.ErrTooManyRequests, "Too many failed attempts. Please wait a few minutes.")
)
