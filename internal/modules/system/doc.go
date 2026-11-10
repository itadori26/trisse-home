// Package system shows host metrics (CPU, memory, disks, load, uptime,
// temperature) read from the host's /proc and /sys, which are mounted
// read-only into the container.
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package system
