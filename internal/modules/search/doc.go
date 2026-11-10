// Package search resolves queries from the search bar to a search provider
// URL. DuckDuckGo is the default, so its !bangs work out of the box; custom
// shortcuts can override them. Providers are interchangeable strategies.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package search
