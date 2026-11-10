// Package rss is a small feed reader: subscribed RSS/Atom feeds are fetched
// in the background and their items can be marked as read.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package rss
