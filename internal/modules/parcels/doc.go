// Package parcels tracks parcels: the user adds tracking numbers, a
// background job asks the carrier's API for the latest status.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package parcels
