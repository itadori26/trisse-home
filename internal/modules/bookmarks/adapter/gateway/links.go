package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/sqlutil"
)

// LoadSettings implements usecase.SettingsRepository; defaults are off.
func (r *Repository) LoadSettings(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	err := r.db.QueryRowContext(ctx, `SELECT enrich, check_links FROM bookmark_settings WHERE id = 1`).Scan(&s.Enrich, &s.CheckLinks)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	return s, nil
}

// SaveSettings implements usecase.SettingsRepository.
func (r *Repository) SaveSettings(ctx context.Context, s domain.Settings) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO bookmark_settings (id, enrich, check_links) VALUES (1, ?, ?)
		ON CONFLICT (id) DO UPDATE SET enrich = excluded.enrich, check_links = excluded.check_links`, s.Enrich, s.CheckLinks)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

// ListLinkStatuses implements usecase.LinkStatusRepository.
func (r *Repository) ListLinkStatuses(ctx context.Context) (map[string]domain.LinkStatus, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT bookmark_id, state, code, target, checked_at FROM bookmark_links`)
	if err != nil {
		return nil, fmt.Errorf("query link statuses: %w", err)
	}
	defer rows.Close()
	out := map[string]domain.LinkStatus{}
	for rows.Next() {
		var id, state, target, checked string
		var code int
		if err := rows.Scan(&id, &state, &code, &target, &checked); err != nil {
			return nil, fmt.Errorf("scan link status: %w", err)
		}
		at, err := sqlutil.ParseTime(checked)
		if err != nil {
			return nil, err
		}
		out[id] = domain.LinkStatus{State: domain.LinkState(state), Code: code, Target: target, CheckedAt: at}
	}
	return out, rows.Err()
}

// SaveLinkStatus implements usecase.LinkStatusRepository.
func (r *Repository) SaveLinkStatus(ctx context.Context, id string, s domain.LinkStatus) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO bookmark_links (bookmark_id, state, code, target, checked_at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (bookmark_id) DO UPDATE SET state = excluded.state, code = excluded.code,
			target = excluded.target, checked_at = excluded.checked_at`,
		id, string(s.State), s.Code, s.Target, sqlutil.FormatTime(s.CheckedAt))
	if err != nil {
		return fmt.Errorf("upsert link status: %w", err)
	}
	return nil
}

// SaveIcon implements usecase.IconStore.
func (r *Repository) SaveIcon(ctx context.Context, id string, icon domain.Icon) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO bookmark_icons (bookmark_id, content_type, data) VALUES (?, ?, ?)
		ON CONFLICT (bookmark_id) DO UPDATE SET content_type = excluded.content_type, data = excluded.data`,
		id, icon.ContentType, icon.Data)
	if err != nil {
		return fmt.Errorf("upsert icon: %w", err)
	}
	return nil
}

// Icon implements usecase.IconStore.
func (r *Repository) Icon(ctx context.Context, id string) (domain.Icon, error) {
	var icon domain.Icon
	err := r.db.QueryRowContext(ctx, `SELECT content_type, data FROM bookmark_icons WHERE bookmark_id = ?`, id).Scan(&icon.ContentType, &icon.Data)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Icon{}, domain.ErrIconNotFound
	}
	if err != nil {
		return domain.Icon{}, fmt.Errorf("select icon: %w", err)
	}
	return icon, nil
}

// IconIDs implements usecase.IconStore.
func (r *Repository) IconIDs(ctx context.Context) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT bookmark_id FROM bookmark_icons`)
	if err != nil {
		return nil, fmt.Errorf("query icons: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan icon: %w", err)
		}
		out[id] = true
	}
	return out, rows.Err()
}
