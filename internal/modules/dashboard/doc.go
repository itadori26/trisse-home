// Package dashboard manages the tabs of the start page and the widgets placed
// on a 12-column grid within each tab. Widget content is provided by other
// modules; this module only stores kind, position and opaque configuration.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package dashboard
