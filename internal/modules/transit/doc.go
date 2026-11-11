// Package transit shows the next public transport departures of a stop,
// using a MOTIS API (default: the community service Transitous).
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package transit
