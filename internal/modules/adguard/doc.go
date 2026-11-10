// Package adguard shows statistics of an AdGuard Home instance. Its URL and
// credentials are configured in the UI and stored encrypted.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package adguard
