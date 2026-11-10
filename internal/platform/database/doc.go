// Package database opens the SQLite database and applies embedded,
// per-module SQL migrations.
//
// Ring: Frameworks & Drivers. This is the only package that imports the
// SQLite driver; gateways only see *sql.DB from the standard library.
package database
