// Package wireguard lists WireGuard peers with name, last handshake and
// transferred data. The data comes from a file written by
// scripts/wireguard-export.sh on the host (no secrets, no Docker access).
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package wireguard
