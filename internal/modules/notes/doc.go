// Package notes stores Markdown notes that can be pinned to the dashboard.
// Markdown is rendered and sanitised in the browser; the backend stores text.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package notes
