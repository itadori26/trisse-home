// Package sdnotify implements the small part of the systemd notification
// protocol the native service needs (ADR 0015): READY=1 once the HTTP server
// listens, STOPPING=1 on shutdown and WATCHDOG=1 keep-alive pings while the
// application is healthy. It uses only the standard library and does nothing
// when NOTIFY_SOCKET is unset, so Docker and development are unaffected.
//
// Ring: Frameworks & Drivers.
package sdnotify
