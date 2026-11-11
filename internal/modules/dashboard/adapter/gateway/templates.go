package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"trisse-home/internal/modules/dashboard/domain"
)

type templateWidgetRow struct {
	Kind   string          `json:"kind"`
	X      int             `json:"x"`
	Y      int             `json:"y"`
	W      int             `json:"w"`
	H      int             `json:"h"`
	Config json.RawMessage `json:"config"`
}

func scanTemplate(sc interface{ Scan(...any) error }) (domain.Template, error) {
	var t domain.Template
	var raw string
	if err := sc.Scan(&t.ID, &t.Name, &raw); err != nil {
		return t, err
	}
	var rows []templateWidgetRow
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return t, fmt.Errorf("decode template widgets: %w", err)
	}
	for _, r := range rows {
		t.Widgets = append(t.Widgets, domain.TemplateWidget{Kind: r.Kind, Rect: domain.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}, Config: string(r.Config)})
	}
	return t, nil
}

// ListTemplates implements usecase.TemplateRepository.
func (r *Repository) ListTemplates(ctx context.Context) ([]domain.Template, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, widgets FROM dashboard_templates ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()
	var out []domain.Template
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetTemplate implements usecase.TemplateRepository.
func (r *Repository) GetTemplate(ctx context.Context, id string) (domain.Template, error) {
	t, err := scanTemplate(r.db.QueryRowContext(ctx, `SELECT id, name, widgets FROM dashboard_templates WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Template{}, domain.ErrTemplateNotFound
	}
	if err != nil {
		return domain.Template{}, fmt.Errorf("select template: %w", err)
	}
	return t, nil
}

// CreateTemplate implements usecase.TemplateRepository.
func (r *Repository) CreateTemplate(ctx context.Context, t domain.Template) error {
	rows := make([]templateWidgetRow, 0, len(t.Widgets))
	for _, w := range t.Widgets {
		rows = append(rows, templateWidgetRow{Kind: w.Kind, X: w.Rect.X, Y: w.Rect.Y, W: w.Rect.W, H: w.Rect.H, Config: json.RawMessage(w.Config)})
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return fmt.Errorf("encode template widgets: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `INSERT INTO dashboard_templates (id, name, widgets) VALUES (?, ?, ?)`, t.ID, t.Name, string(raw)); err != nil {
		return fmt.Errorf("insert template: %w", err)
	}
	return nil
}

// DeleteTemplate implements usecase.TemplateRepository.
func (r *Repository) DeleteTemplate(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM dashboard_templates WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	return nil
}

// CreateTabWithWidgets implements usecase.TemplateRepository atomically.
func (r *Repository) CreateTabWithWidgets(ctx context.Context, t domain.Tab, widgets []domain.Widget) error {
	return r.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO dashboard_tabs (id, name, position, background) VALUES (?, ?, ?, ?)`,
			t.ID(), t.Name(), t.Position(), t.Background()); err != nil {
			return fmt.Errorf("insert tab: %w", err)
		}
		for _, w := range widgets {
			rc := w.Rect()
			if _, err := tx.ExecContext(ctx, `INSERT INTO dashboard_widgets (`+widgetColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				w.ID(), w.TabID(), w.Kind(), rc.X, rc.Y, rc.W, rc.H, w.Config()); err != nil {
				return fmt.Errorf("insert widget: %w", err)
			}
		}
		return nil
	})
}
