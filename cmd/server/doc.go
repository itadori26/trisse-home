// Command server is the composition root of trisse-home. It reads the
// configuration, builds the platform services, wires every enabled module via
// constructor injection and starts the HTTP server and background jobs. It
// contains no business logic.
//
// Ring: Frameworks & Drivers (composition root).
package main
