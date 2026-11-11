// Package nextcloud shows the status of a Nextcloud instance: version,
// maintenance mode, the quota of the configured account and, for admins,
// server information from the serverinfo app.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package nextcloud
