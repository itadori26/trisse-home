package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
	"waylight/internal/shared/sqlutil"
)

// Repository stores tabs and widgets in SQLite and implements both
// usecase.TabRepository and usecase.WidgetRepository.
type Repository struct {
	db     *sql.DB
	shared func(context.Context) bool
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// WithSharedTabs makes tabs that others shared with the principal visible
// while on(ctx) says so (team tabs, #295). Without it every account sees only
// its own tabs, so a build without ee/ never shows a shared tab, and neither
// does an instance whose licence no longer covers team tabs.
func (r *Repository) WithSharedTabs(on func(context.Context) bool) *Repository {
	r.shared = on
	return r
}

// visibleTabs is the condition for the tabs the principal of ctx may see
// (ADR 0018): guests and kiosk displays see the tabs shared with them by ID,
// everyone else their own tabs and, once team tabs exist (#295), those
// shared with their groups. Background jobs see all.
func (r *Repository) visibleTabs(ctx context.Context) (string, []any) {
	p := kernel.PrincipalFrom(ctx)
	if p.TabScoped() {
		if len(p.Tabs) == 0 {
			return "0 = 1", nil
		}
		args := make([]any, len(p.Tabs))
		for i, t := range p.Tabs {
			args[i] = t
		}
		return "id IN (?" + strings.Repeat(", ?", len(p.Tabs)-1) + ")", args
	}
	if r.shared == nil || !r.shared(ctx) {
		return sqlutil.OwnedBy(p, "")
	}
	return sqlutil.VisibleTo(p, "")
}

// ownedTabs is the condition for the tabs the principal of ctx may change:
// only its own, whatever their visibility.
func ownedTabs(ctx context.Context) (string, []any) {
	return sqlutil.OwnedBy(kernel.PrincipalFrom(ctx), "")
}

// ListTabs returns the tabs the principal may see, ordered by position.
func (r *Repository) ListTabs(ctx context.Context) ([]domain.Tab, error) {
	cond, args := r.visibleTabs(ctx)
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, position, background, rules, mobile_layout, zones, sample, owner_id, visibility FROM dashboard_tabs WHERE `+cond+` ORDER BY position, name`, args...) //nolint:gosec // cond comes from sqlutil with bound arguments
	if err != nil {
		return nil, fmt.Errorf("query tabs: %w", err)
	}
	defer rows.Close()
	var tabs []domain.Tab
	for rows.Next() {
		var id, name, bg, rules, owner, vis string
		var mobile sql.NullString
		var pos int
		var zoned, sample bool
		if err := rows.Scan(&id, &name, &pos, &bg, &rules, &mobile, &zoned, &sample, &owner, &vis); err != nil {
			return nil, fmt.Errorf("scan tab: %w", err)
		}
		tabs = append(tabs, domain.RestoreTab(id, name, pos, bg).WithRules(decodeRules(rules)).WithMobile(decodeMobile(mobile)).
			WithZones(zoned).WithSample(sample).WithAccess(kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}))
	}
	return tabs, rows.Err()
}

// GetTab returns one tab or domain.ErrTabNotFound.
func (r *Repository) GetTab(ctx context.Context, id string) (domain.Tab, error) {
	var name, bg, rules, owner, vis string
	var mobile sql.NullString
	var pos int
	var zoned, sample bool
	cond, args := r.visibleTabs(ctx)
	err := r.db.QueryRowContext(ctx, `SELECT name, position, background, rules, mobile_layout, zones, sample, owner_id, visibility FROM dashboard_tabs WHERE id = ? AND `+cond, append([]any{id}, args...)...).
		Scan(&name, &pos, &bg, &rules, &mobile, &zoned, &sample, &owner, &vis)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tab{}, domain.ErrTabNotFound
	}
	if err != nil {
		return domain.Tab{}, fmt.Errorf("select tab: %w", err)
	}
	return domain.RestoreTab(id, name, pos, bg).WithRules(decodeRules(rules)).WithMobile(decodeMobile(mobile)).
		WithZones(zoned).WithSample(sample).WithAccess(kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}), nil
}

// OwnTabs lists the principal's own tabs with their visibility, for sharing
// them as team tabs (#295).
func (r *Repository) OwnTabs(ctx context.Context) ([]domain.Tab, error) {
	cond, args := ownedTabs(ctx)
	//nolint:gosec // cond comes from sqlutil with bound arguments
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, position, background, owner_id, visibility FROM dashboard_tabs WHERE `+cond+` ORDER BY position, name`, args...)
	if err != nil {
		return nil, fmt.Errorf("query own tabs: %w", err)
	}
	defer rows.Close()
	var out []domain.Tab
	for rows.Next() {
		var id, name, bg, owner, vis string
		var pos int
		if err := rows.Scan(&id, &name, &pos, &bg, &owner, &vis); err != nil {
			return nil, fmt.Errorf("scan tab: %w", err)
		}
		out = append(out, domain.RestoreTab(id, name, pos, bg).WithAccess(kernel.Access{Owner: owner, Visibility: kernel.Visibility(vis)}))
	}
	return out, rows.Err()
}

// SetTabVisibility shares an own tab (#295); domain.ErrTabNotFound for a tab
// that does not exist or belongs to someone else.
func (r *Repository) SetTabVisibility(ctx context.Context, id string, v kernel.Visibility) error {
	cond, args := ownedTabs(ctx)
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `UPDATE dashboard_tabs SET visibility = ? WHERE id = ? AND `+cond, append([]any{string(v), id}, args...)...)
	if err != nil {
		return fmt.Errorf("update tab visibility: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrTabNotFound
	}
	return nil
}

// CreateTab inserts a private tab owned by the principal of ctx.
func (r *Repository) CreateTab(ctx context.Context, t domain.Tab) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO dashboard_tabs (id, name, position, background, rules, mobile_layout, zones, sample, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID(), t.Name(), t.Position(), t.Background(), encodeRules(t.Rules()), encodeMobile(t.Mobile()), t.Zoned(), t.Sample(),
		kernel.PrincipalFrom(ctx).AccountID, string(kernel.Private))
	if err != nil {
		return fmt.Errorf("insert tab: %w", err)
	}
	return nil
}

// UpdateTab stores a tab's settings; owner and visibility stay.
func (r *Repository) UpdateTab(ctx context.Context, t domain.Tab) error {
	cond, args := ownedTabs(ctx)
	res, err := r.db.ExecContext(ctx, `UPDATE dashboard_tabs SET name = ?, position = ?, background = ?, rules = ?, mobile_layout = ?, zones = ?, sample = ? WHERE id = ? AND `+cond, //nolint:gosec // cond comes from sqlutil with bound arguments
		append([]any{t.Name(), t.Position(), t.Background(), encodeRules(t.Rules()), encodeMobile(t.Mobile()), t.Zoned(), t.Sample(), t.ID()}, args...)...)
	if err != nil {
		return fmt.Errorf("update tab: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrTabNotFound
	}
	return nil
}

// DeleteTab removes an own tab; widgets are removed by the foreign key
// cascade.
func (r *Repository) DeleteTab(ctx context.Context, id string) error {
	cond, args := ownedTabs(ctx)
	if _, err := r.db.ExecContext(ctx, `DELETE FROM dashboard_tabs WHERE id = ? AND `+cond, append([]any{id}, args...)...); err != nil { //nolint:gosec // cond comes from sqlutil with bound arguments
		return fmt.Errorf("delete tab: %w", err)
	}
	return nil
}

// SaveTabOrder stores the positions of all given tabs in one transaction.
func (r *Repository) SaveTabOrder(ctx context.Context, tabs []domain.Tab) error {
	cond, args := ownedTabs(ctx)
	return r.inTx(ctx, func(tx *sql.Tx) error {
		for _, t := range tabs {
			if _, err := tx.ExecContext(ctx, `UPDATE dashboard_tabs SET position = ? WHERE id = ? AND `+cond, append([]any{t.Position(), t.ID()}, args...)...); err != nil { //nolint:gosec // cond comes from sqlutil with bound arguments
				return fmt.Errorf("update tab position: %w", err)
			}
		}
		return nil
	})
}

const widgetColumns = `id, tab_id, kind, x, y, w, h, config, size, zone`

func scanWidget(s interface{ Scan(...any) error }) (domain.Widget, error) {
	var id, tabID, kind, config, size, zone string
	var rect domain.Rect
	if err := s.Scan(&id, &tabID, &kind, &rect.X, &rect.Y, &rect.W, &rect.H, &config, &size, &zone); err != nil {
		return domain.Widget{}, err
	}
	return domain.RestoreWidget(id, tabID, kind, rect, config, domain.Size(size)).WithZone(domain.Zone(zone)), nil
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

// Widgets have no owner of their own: they belong to their tab, so every
// widget query is limited to tabs the principal may see or change.

// ListWidgets returns the widgets on visible tabs ordered by position.
func (r *Repository) ListWidgets(ctx context.Context) ([]domain.Widget, error) {
	cond, args := r.visibleTabs(ctx)
	return r.queryWidgets(ctx, `SELECT `+widgetColumns+` FROM dashboard_widgets WHERE tab_id IN (SELECT id FROM dashboard_tabs WHERE `+cond+`) ORDER BY y, x`, args...)
}

// ListWidgetsByTab returns the widgets of one visible tab.
func (r *Repository) ListWidgetsByTab(ctx context.Context, tabID string) ([]domain.Widget, error) {
	cond, args := r.visibleTabs(ctx)
	return r.queryWidgets(ctx, `SELECT `+widgetColumns+` FROM dashboard_widgets WHERE tab_id = ? AND tab_id IN (SELECT id FROM dashboard_tabs WHERE `+cond+`) ORDER BY y, x`,
		append([]any{tabID}, args...)...)
}

// GetWidget returns one widget on a visible tab or domain.ErrWidgetNotFound.
func (r *Repository) GetWidget(ctx context.Context, id string) (domain.Widget, error) {
	cond, args := r.visibleTabs(ctx)
	w, err := scanWidget(r.db.QueryRowContext(ctx, `SELECT `+widgetColumns+` FROM dashboard_widgets WHERE id = ? AND tab_id IN (SELECT id FROM dashboard_tabs WHERE `+cond+`)`,
		append([]any{id}, args...)...))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Widget{}, domain.ErrWidgetNotFound
	}
	if err != nil {
		return domain.Widget{}, fmt.Errorf("select widget: %w", err)
	}
	return w, nil
}

// CreateWidget inserts a widget on an own tab; on any other tab it returns
// domain.ErrTabNotFound.
func (r *Repository) CreateWidget(ctx context.Context, w domain.Widget) error {
	rect := w.Rect()
	cond, args := ownedTabs(ctx)
	//nolint:gosec // cond comes from sqlutil with bound arguments
	res, err := r.db.ExecContext(ctx, `INSERT INTO dashboard_widgets (`+widgetColumns+`) SELECT ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		WHERE EXISTS (SELECT 1 FROM dashboard_tabs WHERE id = ? AND `+cond+`)`,
		append([]any{w.ID(), w.TabID(), w.Kind(), rect.X, rect.Y, rect.W, rect.H, w.Config(), string(w.Size()), string(w.Zone()), w.TabID()}, args...)...)
	if err != nil {
		return fmt.Errorf("insert widget: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrTabNotFound
	}
	return nil
}

// UpdateWidget stores the configuration, rectangle and size of a widget.
func (r *Repository) UpdateWidget(ctx context.Context, w domain.Widget) error {
	rect := w.Rect()
	cond, args := ownedTabs(ctx)
	_, err := r.db.ExecContext(ctx, `UPDATE dashboard_widgets SET x = ?, y = ?, w = ?, h = ?, config = ?, size = ?, zone = ? WHERE id = ? AND tab_id IN (SELECT id FROM dashboard_tabs WHERE `+cond+`)`, //nolint:gosec // cond comes from sqlutil with bound arguments
		append([]any{rect.X, rect.Y, rect.W, rect.H, w.Config(), string(w.Size()), string(w.Zone()), w.ID()}, args...)...)
	if err != nil {
		return fmt.Errorf("update widget: %w", err)
	}
	return nil
}

// SaveWidgetRects stores the rectangles and sizes of several widgets atomically.
func (r *Repository) SaveWidgetRects(ctx context.Context, widgets []domain.Widget) error {
	cond, args := ownedTabs(ctx)
	return r.inTx(ctx, func(tx *sql.Tx) error {
		for _, w := range widgets {
			rect := w.Rect()
			if _, err := tx.ExecContext(ctx, `UPDATE dashboard_widgets SET x = ?, y = ?, w = ?, h = ?, size = ?, zone = ? WHERE id = ? AND tab_id IN (SELECT id FROM dashboard_tabs WHERE `+cond+`)`, //nolint:gosec // cond comes from sqlutil with bound arguments
				append([]any{rect.X, rect.Y, rect.W, rect.H, string(w.Size()), string(w.Zone()), w.ID()}, args...)...); err != nil {
				return fmt.Errorf("update widget rect: %w", err)
			}
		}
		return nil
	})
}

// DeleteWidget removes a widget from an own tab.
func (r *Repository) DeleteWidget(ctx context.Context, id string) error {
	cond, args := ownedTabs(ctx)
	if _, err := r.db.ExecContext(ctx, `DELETE FROM dashboard_widgets WHERE id = ? AND tab_id IN (SELECT id FROM dashboard_tabs WHERE `+cond+`)`, //nolint:gosec // cond comes from sqlutil with bound arguments
		append([]any{id}, args...)...); err != nil {
		return fmt.Errorf("delete widget: %w", err)
	}
	return nil
}

// ReplaceAll implements usecase.ConfigRepository in one transaction. It
// works on the tabs of the principal of ctx (config as code passes the
// admin, ADR 0018): kept tabs lose their old widgets; removed tabs take
// theirs along (cascade). Other accounts' tabs are not touched.
func (r *Repository) ReplaceAll(ctx context.Context, tabs []domain.Tab, widgets []domain.Widget) error {
	cond, args := ownedTabs(ctx)
	owner := kernel.PrincipalFrom(ctx).AccountID
	return r.inTx(ctx, func(tx *sql.Tx) error {
		keep := map[string]bool{}
		for _, t := range tabs {
			keep[t.ID()] = true
		}
		existing, err := queryIDs(ctx, tx, `SELECT id FROM dashboard_tabs WHERE `+cond, args...)
		if err != nil {
			return fmt.Errorf("list tabs: %w", err)
		}
		for _, id := range existing {
			if keep[id] {
				continue
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM dashboard_tabs WHERE id = ?`, id); err != nil {
				return fmt.Errorf("delete tab: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM dashboard_widgets WHERE tab_id IN (SELECT id FROM dashboard_tabs WHERE `+cond+`)`, args...); err != nil { //nolint:gosec // cond comes from sqlutil with bound arguments
			return fmt.Errorf("clear widgets: %w", err)
		}
		for _, t := range tabs {
			// The upsert only takes over a tab of the same owner; an ID of
			// someone else's tab is refused instead of moved. A tab from the
			// config file is no example tab any more.
			res, err := tx.ExecContext(ctx, `INSERT INTO dashboard_tabs (id, name, position, background, rules, mobile_layout, zones, owner_id, visibility) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT (id) DO UPDATE SET name = excluded.name, position = excluded.position, background = excluded.background,
				rules = excluded.rules, mobile_layout = excluded.mobile_layout, zones = excluded.zones, sample = 0 WHERE dashboard_tabs.owner_id = excluded.owner_id`,
				t.ID(), t.Name(), t.Position(), t.Background(), encodeRules(t.Rules()), encodeMobile(t.Mobile()), t.Zoned(), owner, string(kernel.Private))
			if err != nil {
				return fmt.Errorf("upsert tab: %w", err)
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return domain.ErrTabNotFound
			}
		}
		for _, w := range widgets {
			rc := w.Rect()
			if _, err := tx.ExecContext(ctx, `INSERT INTO dashboard_widgets (`+widgetColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				w.ID(), w.TabID(), w.Kind(), rc.X, rc.Y, rc.W, rc.H, w.Config(), string(w.Size()), string(w.Zone())); err != nil {
				return fmt.Errorf("insert widget: %w", err)
			}
		}
		return nil
	})
}

// queryIDs runs a query that selects one id column.
func queryIDs(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]string, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
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
