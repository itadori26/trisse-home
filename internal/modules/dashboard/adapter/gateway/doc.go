// Package gateway implements the dashboard repositories on SQLite.
//
// Ring: Interface Adapters.
package gateway

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}
