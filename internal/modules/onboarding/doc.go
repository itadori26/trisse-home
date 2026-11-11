// Package onboarding remembers whether the guided tour of the web interface
// has been completed, so it is shown exactly once per account.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package onboarding
