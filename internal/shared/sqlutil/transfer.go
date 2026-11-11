package sqlutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Table describes a table that takes part in a data export: its rows are
// exported except the Exclude columns (secrets, caches), which keep their
// database default on import.
type Table struct {
	Name    string
	Exclude []string
}

// Rows are the exported rows of one table (column → value).
type Rows []map[string]any

var identifier = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// ErrInvalidExport is returned for malformed import data.
var ErrInvalidExport = errors.New("invalid export data")

func columns(ctx context.Context, q interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, table string) ([]string, error) {
	if !identifier.MatchString(table) {
		return nil, fmt.Errorf("invalid table name %q", table)
	}
	rows, err := q.QueryContext(ctx, `SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return nil, fmt.Errorf("columns of %s: %w", table, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("unknown table %s", table)
	}
	return out, rows.Err()
}

// DumpTables reads the given tables in order.
func DumpTables(ctx context.Context, db *sql.DB, tables []Table) (map[string]Rows, error) {
	out := map[string]Rows{}
	for _, t := range tables {
		rows, err := dumpTable(ctx, db, t)
		if err != nil {
			return nil, err
		}
		out[t.Name] = rows
	}
	return out, nil
}

func dumpTable(ctx context.Context, db *sql.DB, t Table) (Rows, error) {
	cols, err := columns(ctx, db, t.Name)
	if err != nil {
		return nil, err
	}
	cols = slices.DeleteFunc(cols, func(c string) bool { return slices.Contains(t.Exclude, c) })
	rows, err := db.QueryContext(ctx, `SELECT `+strings.Join(cols, ", ")+` FROM `+t.Name+` ORDER BY rowid`) //nolint:gosec // identifiers come from pragma_table_info and are validated
	if err != nil {
		return nil, fmt.Errorf("dump %s: %w", t.Name, err)
	}
	defer rows.Close()
	list := Rows{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scan %s: %w", t.Name, err)
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			if b, ok := vals[i].([]byte); ok {
				vals[i] = string(b)
			}
			row[c] = vals[i]
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

// RestoreTables replaces the content of the given tables with data in one
// transaction: rows are deleted in reverse order, inserted in order (only
// known, non-excluded columns are used) and foreign keys are checked before
// committing. Tables missing in data are emptied.
func RestoreTables(ctx context.Context, db *sql.DB, tables []Table, data map[string]Rows) error {
	for name := range data {
		if !slices.ContainsFunc(tables, func(t Table) bool { return t.Name == name }) {
			return fmt.Errorf("%w: unexpected table %q", ErrInvalidExport, name)
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for i := len(tables) - 1; i >= 0; i-- {
		if _, err := tx.ExecContext(ctx, `DELETE FROM `+tables[i].Name); err != nil { //nolint:gosec // validated identifier
			return fmt.Errorf("clear %s: %w", tables[i].Name, err)
		}
	}
	for _, t := range tables {
		known, err := columns(ctx, tx, t.Name)
		if err != nil {
			return err
		}
		for n, row := range data[t.Name] {
			var cols []string
			var args []any
			for _, c := range known {
				v, ok := row[c]
				if !ok || slices.Contains(t.Exclude, c) {
					continue
				}
				switch x := v.(type) {
				case nil, string, bool, float64, int64:
				case json.Number:
					if i, err := x.Int64(); err == nil {
						v = i
					} else if f, err := x.Float64(); err == nil {
						v = f
					}
				default:
					return fmt.Errorf("%w: %s row %d column %s has an unsupported value", ErrInvalidExport, t.Name, n+1, c)
				}
				cols, args = append(cols, c), append(args, v)
			}
			if len(cols) == 0 {
				return fmt.Errorf("%w: %s row %d is empty", ErrInvalidExport, t.Name, n+1)
			}
			q := `INSERT INTO ` + t.Name + ` (` + strings.Join(cols, ", ") + `) VALUES (` + strings.TrimSuffix(strings.Repeat("?, ", len(cols)), ", ") + `)` //nolint:gosec // validated identifiers
			if _, err := tx.ExecContext(ctx, q, args...); err != nil {
				return fmt.Errorf("%w: %s row %d: %w", ErrInvalidExport, t.Name, n+1, err)
			}
		}
	}
	if err := checkForeignKeys(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

func checkForeignKeys(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("check foreign keys: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		return fmt.Errorf("%w: references between the rows are broken", ErrInvalidExport)
	}
	return rows.Err()
}
