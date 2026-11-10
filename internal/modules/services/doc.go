// Package services shows tiles for self-hosted services with their online
// status, checked periodically in the background.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package services
