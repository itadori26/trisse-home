package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trisse-home/internal/modules/dashboard/domain"
)

// Repository stores tabs and widgets in SQLite and implements both
// usecase.TabRepository and usecase.WidgetRepository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// ListTabs returns all tabs ordered by position.
func (r *Repository) ListTabs(ctx context.Context) ([]domain.Tab, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, position, background, rules FROM dashboard_tabs ORDER BY position, name`)
	if err != nil {
		return nil, fmt.Errorf("query tabs: %w", err)
	}
	defer rows.Close()
	var tabs []domain.Tab
	for rows.Next() {
		var id, name, bg, rules string
		var pos int
		if err := rows.Scan(&id, &name, &pos, &bg, &rules); err != nil {
			return nil, fmt.Errorf("scan tab: %w", err)
		}
		tabs = append(tabs, domain.RestoreTab(id, name, pos, bg).WithRules(decodeRules(rules)))
	}
	return tabs, rows.Err()
}

// GetTab returns one tab or domain.ErrTabNotFound.
func (r *Repository) GetTab(ctx context.Context, id string) (domain.Tab, error) {
	var name, bg, rules string
	var pos int
	err := r.db.QueryRowContext(ctx, `SELECT name, position, background, rules FROM dashboard_tabs WHERE id = ?`, id).Scan(&name, &pos, &bg, &rules)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tab{}, domain.ErrTabNotFound
	}
	if err != nil {
		return domain.Tab{}, fmt.Errorf("select tab: %w", err)
	}
	return domain.RestoreTab(id, name, pos, bg).WithRules(decodeRules(rules)), nil
}

// CreateTab inserts a tab.
func (r *Repository) CreateTab(ctx context.Context, t domain.Tab) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO dashboard_tabs (id, name, position, background, rules) VALUES (?, ?, ?, ?, ?)`,
		t.ID(), t.Name(), t.Position(), t.Background(), encodeRules(t.Rules()))
	if err != nil {
		return fmt.Errorf("insert tab: %w", err)
	}
	return nil
}

// UpdateTab stores name and position of a tab.
func (r *Repository) UpdateTab(ctx context.Context, t domain.Tab) error {
	res, err := r.db.ExecContext(ctx, `UPDATE dashboard_tabs SET name = ?, position = ?, background = ?, rules = ? WHERE id = ?`,
		t.Name(), t.Position(), t.Background(), encodeRules(t.Rules()), t.ID())
	if err != nil {
		return fmt.Errorf("update tab: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrTabNotFound
	}
	return nil
}

// DeleteTab removes a tab; widgets are removed by the foreign key cascade.
func (r *Repository) DeleteTab(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM dashboard_tabs WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete tab: %w", err)
	}
	return nil
}

// SaveTabOrder stores the positions of all given tabs in one transaction.
func (r *Repository) SaveTabOrder(ctx context.Context, tabs []domain.Tab) error {
	return r.inTx(ctx, func(tx *sql.Tx) error {
		for _, t := range tabs {
			if _, err := tx.ExecContext(ctx, `UPDATE dashboard_tabs SET position = ? WHERE id = ?`, t.Position(), t.ID()); err != nil {
				return fmt.Errorf("update tab position: %w", err)
			}
		}
		return nil
	})
}

const widgetColumns = `id, tab_id, kind, x, y, w, h, config`

func scanWidget(s interface{ Scan(...any) error }) (domain.Widget, error) {
	var id, tabID, kind, config string
	var rect domain.Rect
	if err := s.Scan(&id, &tabID, &kind, &rect.X, &rect.Y, &rect.W, &rect.H, &config); err != nil {
		return domain.Widget{}, err
	}
	return domain.RestoreWidget(id, tabID, kind, rect, config), nil
}

func (r *Repository) queryWidgets(ctx context.Context, query string, args ...any) ([]domain.Widget, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query widgets: %w", err)
	}
	defer rows.Close()
	var out []domain.Widget
	for rows.Next() {
		w, err := scanWidget(rows)
		if err != nil {
			return nil, fmt.Errorf("scan widget: %w", err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ListWidgets returns all widgets ordered by position.
func (r *Repository) ListWidgets(ctx context.Context) ([]domain.Widget, error) {
	return r.queryWidgets(ctx, `SELECT `+widgetColumns+` FROM dashboard_widgets ORDER BY y, x`)
}

// ListWidgetsByTab returns the widgets of one tab.
func (r *Repository) ListWidgetsByTab(ctx context.Context, tabID string) ([]domain.Widget, error) {
	return r.queryWidgets(ctx, `SELECT `+widgetColumns+` FROM dashboard_widgets WHERE tab_id = ? ORDER BY y, x`, tabID)
}

// GetWidget returns one widget or domain.ErrWidgetNotFound.
func (r *Repository) GetWidget(ctx context.Context, id string) (domain.Widget, error) {
	w, err := scanWidget(r.db.QueryRowContext(ctx, `SELECT `+widgetColumns+` FROM dashboard_widgets WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Widget{}, domain.ErrWidgetNotFound
	}
	if err != nil {
		return domain.Widget{}, fmt.Errorf("select widget: %w", err)
	}
	return w, nil
}

// CreateWidget inserts a widget.
func (r *Repository) CreateWidget(ctx context.Context, w domain.Widget) error {
	rect := w.Rect()
	_, err := r.db.ExecContext(ctx, `INSERT INTO dashboard_widgets (`+widgetColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		w.ID(), w.TabID(), w.Kind(), rect.X, rect.Y, rect.W, rect.H, w.Config())
	if err != nil {
		return fmt.Errorf("insert widget: %w", err)
	}
	return nil
}

// UpdateWidget stores the configuration and rectangle of a widget.
func (r *Repository) UpdateWidget(ctx context.Context, w domain.Widget) error {
	rect := w.Rect()
	_, err := r.db.ExecContext(ctx, `UPDATE dashboard_widgets SET x = ?, y = ?, w = ?, h = ?, config = ? WHERE id = ?`,
		rect.X, rect.Y, rect.W, rect.H, w.Config(), w.ID())
	if err != nil {
		return fmt.Errorf("update widget: %w", err)
	}
	return nil
}

// SaveWidgetRects stores the rectangles of several widgets atomically.
func (r *Repository) SaveWidgetRects(ctx context.Context, widgets []domain.Widget) error {
	return r.inTx(ctx, func(tx *sql.Tx) error {
		for _, w := range widgets {
			rect := w.Rect()
			if _, err := tx.ExecContext(ctx, `UPDATE dashboard_widgets SET x = ?, y = ?, w = ?, h = ? WHERE id = ?`,
				rect.X, rect.Y, rect.W, rect.H, w.ID()); err != nil {
				return fmt.Errorf("update widget rect: %w", err)
			}
		}
		return nil
	})
}

// DeleteWidget removes a widget.
func (r *Repository) DeleteWidget(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM dashboard_widgets WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete widget: %w", err)
	}
	return nil
}

func (r *Repository) inTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// KindCatalog is a static usecase.KindCatalog built from the enabled modules.
type KindCatalog map[string]struct{}

// NewKindCatalog returns a catalog containing kinds.
func NewKindCatalog(kinds []string) KindCatalog {
	c := make(KindCatalog, len(kinds))
	for _, k := range kinds {
		c[k] = struct{}{}
	}
	return c
}

// Available implements usecase.KindCatalog.
func (c KindCatalog) Available(kind string) bool {
	_, ok := c[kind]
	return ok
}
