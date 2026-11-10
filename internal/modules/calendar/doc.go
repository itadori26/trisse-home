// Package calendar shows upcoming events from CalDAV calendars (e.g.
// Nextcloud). Credentials are configured in the UI and stored encrypted.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package calendar
