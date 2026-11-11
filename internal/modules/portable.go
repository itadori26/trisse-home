package modules

import (
	"context"
	"database/sql"

	"trisse-home/internal/shared/sqlutil"
)

// Portable is implemented by modules whose user data can be exported and
// imported as JSON (see the transfer module). Exports never contain secrets;
// ImportData replaces the module's data atomically.
type Portable interface {
	ExportData(ctx context.Context) (map[string]sqlutil.Rows, error)
	ImportData(ctx context.Context, data map[string]sqlutil.Rows) error
}

// Transferable implements Portable for modules whose exportable data lives
// in plain tables. Embed it and set DB and Tables (parents before children).
type Transferable struct {
	DB     *sql.DB
	Tables []sqlutil.Table
}

// ExportData implements Portable.
func (t Transferable) ExportData(ctx context.Context) (map[string]sqlutil.Rows, error) {
	return sqlutil.DumpTables(ctx, t.DB, t.Tables)
}

// ImportData implements Portable.
func (t Transferable) ImportData(ctx context.Context, data map[string]sqlutil.Rows) error {
	return sqlutil.RestoreTables(ctx, t.DB, t.Tables, data)
}
