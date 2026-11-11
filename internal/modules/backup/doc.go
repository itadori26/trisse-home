// Package backup creates consistent SQLite backups on a schedule, prunes old
// ones according to a retention policy and offers them for download.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package backup
