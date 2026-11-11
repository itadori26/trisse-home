// Package transfer exports all user data (layout, bookmarks, notes, to-dos,
// feeds, services and settings, never secrets) as one versioned JSON file
// and imports such files again. Each participating module contributes its
// own section via modules.Portable.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package transfer
