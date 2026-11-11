// Package notifications sends push messages via ntfy when monitored things
// change state: a service goes offline, a container becomes unhealthy, a
// disk fills up, a certificate is about to expire or a backup is overdue.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package notifications
