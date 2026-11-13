package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/shared/kernel"
)

type themeRepo struct{ a domain.Appearance }

func (m *themeRepo) Load(context.Context) (domain.Appearance, error) { return m.a, nil }
func (m *themeRepo) Save(_ context.Context, a domain.Appearance) error {
	m.a = a
	return nil
}

type memStore struct {
	imported []domain.ThemeDef
	backup   *StyleBackup
}

func (m *memStore) Imported(context.Context) ([]domain.ThemeDef, error) { return m.imported, nil }
func (m *memStore) SaveImported(_ context.Context, t domain.ThemeDef) error {
	m.imported = append(m.imported, t)
	return nil
}
func (m *memStore) DeleteImported(_ context.Context, id string) error {
	for i, t := range m.imported {
		if t.ID == id {
			m.imported = append(m.imported[:i], m.imported[i+1:]...)
			return nil
		}
	}
	return domain.ErrThemeNotFound
}
func (m *memStore) Backup(context.Context) (StyleBackup, bool, error) {
	if m.backup == nil {
		return StyleBackup{}, false, nil
	}
	return *m.backup, true, nil
}
func (m *memStore) SaveBackup(_ context.Context, b StyleBackup) error {
	m.backup = &b
	return nil
}

type fakeCodec struct{ t domain.ThemeDef }

func (f fakeCodec) Decode(data []byte) (domain.ThemeDef, error) {
	if string(data) != "ok" {
		return domain.ThemeDef{}, errors.New("syntax")
	}
	return f.t, nil
}
func (fakeCodec) Encode(t domain.ThemeDef) ([]byte, error) { return []byte(t.ID), nil }

func theme(t *testing.T, id string, scheme domain.Theme) domain.ThemeDef {
	t.Helper()
	th, err := domain.NewThemeDef(id, id, "", scheme, map[string]string{"--color-bg": "#000", "--color-surface": "#111", "--color-text": "#fff", "--color-accent": "#f0f"})
	if err != nil {
		t.Fatal(err)
	}
	return th
}

func TestThemes(t *testing.T) {
	ctx := context.Background()
	own, _ := domain.NewAppearance(domain.ThemeLight, "#123456", "", "body{font-size:20px}")
	_ = own.SetLanguage("de")
	repo := &themeRepo{a: own}
	store := &memStore{}
	nord := theme(t, "nord", domain.ThemeDark)
	nord.BuiltIn = true
	uc := NewThemes(repo, store, fakeCodec{t: theme(t, "nord", domain.ThemeLight)}, []domain.ThemeDef{nord, theme(t, "latte", domain.ThemeLight)})

	res, err := uc.Apply(ctx, "nord")
	if err != nil || res.Theme != "dark" || res.AccentColor != "" || res.Language != "de" || domain.AppliedTheme(res.CustomCSS) != "nord" ||
		*store.backup != (StyleBackup{Theme: domain.ThemeLight, CSS: "body{font-size:20px}", AccentColor: "#123456"}) {
		t.Fatalf("%+v %v", res, err)
	}
	// A second theme does not overwrite the backup of the user's own CSS.
	_, _ = uc.Apply(ctx, "latte")
	if store.backup.CSS != "body{font-size:20px}" {
		t.Fatal("backup overwritten by theme CSS")
	}
	g, _ := uc.Gallery(ctx)
	if g.Applied != "latte" || !g.CanRestore || len(g.Themes) != 2 {
		t.Fatalf("%+v", g)
	}
	res, _ = uc.Restore(ctx)
	if res.CustomCSS != "body{font-size:20px}" || res.AccentColor != "#123456" || res.Language != "de" || res.Theme != "light" {
		t.Fatal(res)
	}
	if _, err := uc.Restore(ctx); !errors.Is(err, kernel.ErrConflict) {
		t.Fatal("restoring own CSS again must be refused", err)
	}

	imp, err := uc.Import(ctx, []byte("ok"))
	if err != nil || imp.ID != "nord-2" || imp.BuiltIn {
		t.Fatalf("id clash gets a suffix: %+v %v", imp, err)
	}
	if _, err := uc.Import(ctx, []byte("{")); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	if out, _ := uc.Export(ctx, "nord-2"); string(out) != "nord-2" {
		t.Fatal(string(out))
	}
	if err := uc.Delete(ctx, "nord"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("built-in deleted")
	}
	if err := uc.Delete(ctx, "nord-2"); err != nil || len(store.imported) != 0 {
		t.Fatal(err)
	}
	if _, err := uc.Apply(ctx, "nope"); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
}
