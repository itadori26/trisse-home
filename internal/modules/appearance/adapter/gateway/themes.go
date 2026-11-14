package gateway

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/modules/appearance/usecase"
)

// Built-in themes (#111); see themes/README.md at the repository root.
//
//go:embed themes/*.json
var builtinFiles embed.FS

// themeFile is the shareable theme format (JSON).
type themeFile struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Author string            `json:"author,omitempty"`
	Scheme string            `json:"scheme"`
	Tokens map[string]string `json:"tokens"`
}

// MaxThemeBytes bounds an imported theme file.
const MaxThemeBytes = 16 << 10

// DecodeTheme implements usecase.ThemeCodec: strict JSON, unknown fields refused.
func DecodeTheme(data []byte) (domain.ThemeDef, error) {
	if len(data) > MaxThemeBytes {
		return domain.ThemeDef{}, errors.New("theme file too large")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var f themeFile
	if err := dec.Decode(&f); err != nil {
		return domain.ThemeDef{}, fmt.Errorf("decode theme: %w", err)
	}
	return domain.NewThemeDef(f.ID, f.Name, f.Author, domain.Theme(f.Scheme), f.Tokens)
}

// EncodeTheme writes the shareable form of a theme.
func EncodeTheme(t domain.ThemeDef) ([]byte, error) {
	out, err := json.MarshalIndent(themeFile{ID: t.ID, Name: t.Name, Author: t.Author, Scheme: string(t.Scheme), Tokens: t.Tokens}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode theme: %w", err)
	}
	return append(out, '\n'), nil
}

// Codec implements usecase.ThemeCodec.
type Codec struct{}

// Decode implements usecase.ThemeCodec.
func (Codec) Decode(data []byte) (domain.ThemeDef, error) { return DecodeTheme(data) }

// Encode implements usecase.ThemeCodec.
func (Codec) Encode(t domain.ThemeDef) ([]byte, error) { return EncodeTheme(t) }

// BuiltinThemes reads the embedded themes, sorted by name. A broken file
// is a build error caught by the tests, so it is skipped here.
func BuiltinThemes() []domain.ThemeDef {
	var out []domain.ThemeDef
	_ = fs.WalkDir(builtinFiles, "themes", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(p) != ".json" {
			return nil //nolint:nilerr // skip unreadable entries
		}
		data, _ := builtinFiles.ReadFile(p)
		if t, err := DecodeTheme(data); err == nil {
			t.BuiltIn = true
			out = append(out, t)
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ThemeStore implements usecase.ThemeStore: imported themes and the CSS backup.
type ThemeStore struct{ db *sql.DB }

// NewThemeStore returns a store backed by db.
func NewThemeStore(db *sql.DB) *ThemeStore { return &ThemeStore{db: db} }

// Imported implements usecase.ThemeStore.
func (s *ThemeStore) Imported(ctx context.Context) ([]domain.ThemeDef, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT json FROM appearance_themes ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("select themes: %w", err)
	}
	defer rows.Close()
	var out []domain.ThemeDef
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan theme: %w", err)
		}
		if t, err := DecodeTheme([]byte(raw)); err == nil {
			out = append(out, t)
		}
	}
	return out, rows.Err()
}

// SaveImported implements usecase.ThemeStore.
func (s *ThemeStore) SaveImported(ctx context.Context, t domain.ThemeDef) error {
	data, err := EncodeTheme(t)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO appearance_themes (id, json) VALUES (?, ?)
		ON CONFLICT (id) DO UPDATE SET json = excluded.json`, t.ID, string(data))
	if err != nil {
		return fmt.Errorf("upsert theme: %w", err)
	}
	return nil
}

// DeleteImported implements usecase.ThemeStore.
func (s *ThemeStore) DeleteImported(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM appearance_themes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete theme: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrThemeNotFound
	}
	return nil
}

// Backup implements usecase.ThemeStore.
func (s *ThemeStore) Backup(ctx context.Context) (usecase.StyleBackup, bool, error) {
	var b usecase.StyleBackup
	var theme string
	err := s.db.QueryRowContext(ctx, `SELECT theme, css, accent_color FROM appearance_css_backup WHERE id = 1`).Scan(&theme, &b.CSS, &b.AccentColor)
	if errors.Is(err, sql.ErrNoRows) {
		return usecase.StyleBackup{}, false, nil
	}
	if err != nil {
		return usecase.StyleBackup{}, false, fmt.Errorf("select backup: %w", err)
	}
	b.Theme = domain.Theme(theme)
	return b, true, nil
}

// SaveBackup implements usecase.ThemeStore.
func (s *ThemeStore) SaveBackup(ctx context.Context, b usecase.StyleBackup) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO appearance_css_backup (id, theme, css, accent_color) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET theme = excluded.theme, css = excluded.css, accent_color = excluded.accent_color`, string(b.Theme), b.CSS, b.AccentColor)
	if err != nil {
		return fmt.Errorf("upsert backup: %w", err)
	}
	return nil
}
