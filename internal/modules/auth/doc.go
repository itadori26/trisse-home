// Package auth is the reference module: single-user account setup, login,
// logout, session validation and CSRF protection, implemented through all four
// rings of the clean architecture. module.go wires the rings together.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package auth
