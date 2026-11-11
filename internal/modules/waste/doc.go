// Package waste shows the next waste collection dates from an ICS
// subscription (as published by many municipalities) and highlights the
// collection due tomorrow.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package waste
