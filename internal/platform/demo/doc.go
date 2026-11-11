// Package demo provides the infrastructure for DEMO_MODE, a public showcase
// instance with sample data: an HTTP transport that serves fixtures instead
// of reaching the internet, a secret store that refuses to store secrets, a
// middleware that blocks account-changing routes and a start-up guard that
// refuses to run on a database with real data.
//
// The use cases do not know about demo mode: cmd/server swaps these adapters
// in at wiring time.
//
// Ring: Frameworks & Drivers.
package demo
