// Package fritzbox shows the status of an AVM FRITZ!Box: WAN connection,
// external IP and current throughput (via UPnP/IGD, no login needed) and,
// with a FRITZ!Box user, the recent calls and the link speed (via TR-064).
//
// Ring: module wiring. module.go is a small composition root that connects
// the domain, use case and adapter packages below and implements
// modules.Module; it contains no business logic.
package fritzbox
