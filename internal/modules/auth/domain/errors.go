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

// Errors of guest accounts.
var (
	ErrOwnerNotAGuest  = kernel.NewPublicError(kernel.ErrInvalid, "The owner account is not a guest.")
	ErrGuestNotFound   = kernel.NewPublicError(kernel.ErrNotFound, "The guest does not exist.")
	ErrGuestNameTaken  = kernel.NewPublicError(kernel.ErrConflict, "A guest with this name already exists.")
	ErrTooManyGuests   = kernel.NewPublicError(kernel.ErrInvalid, "There are already 20 guests.")
	ErrAccountDisabled = kernel.NewPublicError(kernel.ErrForbidden, "This account is disabled.")
	ErrOwnerOnly       = kernel.NewPublicError(kernel.ErrForbidden, "Only the owner can do this.")
	ErrReadOnly        = kernel.NewPublicError(kernel.ErrForbidden, "Guests can only view this dashboard.")
)
