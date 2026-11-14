package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/shared/kernel"
)

// StyleBackup is the user's own style that applying a theme replaced.
type StyleBackup struct {
	Theme       domain.Theme
	CSS         string
	AccentColor string
}

// ThemeStore keeps imported themes and the style a theme replaced (output port).
type ThemeStore interface {
	Imported(ctx context.Context) ([]domain.ThemeDef, error)
	SaveImported(ctx context.Context, t domain.ThemeDef) error
	DeleteImported(ctx context.Context, id string) error
	Backup(ctx context.Context) (b StyleBackup, ok bool, err error)
	SaveBackup(ctx context.Context, b StyleBackup) error
}

// ThemeCodec reads and writes the shareable theme format (output port).
type ThemeCodec interface {
	Decode(data []byte) (domain.ThemeDef, error)
	Encode(t domain.ThemeDef) ([]byte, error)
}

// Gallery is the overview of the theme gallery.
type Gallery struct {
	Themes []domain.ThemeDef
	// Applied is the id of the theme in use ("" = own or no custom CSS).
	Applied string
	// CanRestore is true when custom CSS was replaced by a theme and is kept.
	CanRestore bool
}

// Themes is the input port of the theme gallery (#111).
type Themes interface {
	Gallery(ctx context.Context) (Gallery, error)
	// Apply writes the theme's tokens as custom CSS; own custom CSS is kept
	// as a backup first.
	Apply(ctx context.Context, id string) (AppearanceResponse, error)
	// Restore puts the backed-up scheme, custom CSS and accent color back.
	Restore(ctx context.Context) (AppearanceResponse, error)
	Import(ctx context.Context, data []byte) (domain.ThemeDef, error)
	Export(ctx context.Context, id string) ([]byte, error)
	Delete(ctx context.Context, id string) error
}

type themes struct {
	repo    Repository
	store   ThemeStore
	codec   ThemeCodec
	builtin []domain.ThemeDef
}

// NewThemes returns the gallery use cases.
func NewThemes(repo Repository, store ThemeStore, codec ThemeCodec, builtin []domain.ThemeDef) Themes {
	return &themes{repo: repo, store: store, codec: codec, builtin: builtin}
}

func (uc *themes) all(ctx context.Context) ([]domain.ThemeDef, error) {
	imported, err := uc.store.Imported(ctx)
	if err != nil {
		return nil, err
	}
	return append(slices.Clone(uc.builtin), imported...), nil
}

func (uc *themes) find(ctx context.Context, id string) (domain.ThemeDef, error) {
	all, err := uc.all(ctx)
	if err != nil {
		return domain.ThemeDef{}, err
	}
	if i := slices.IndexFunc(all, func(t domain.ThemeDef) bool { return t.ID == id }); i >= 0 {
		return all[i], nil
	}
	return domain.ThemeDef{}, domain.ErrThemeNotFound
}

func (uc *themes) Gallery(ctx context.Context) (Gallery, error) {
	all, err := uc.all(ctx)
	if err != nil {
		return Gallery{}, err
	}
	a, err := uc.repo.Load(ctx)
	if err != nil {
		return Gallery{}, fmt.Errorf("load appearance: %w", err)
	}
	_, hasBackup, err := uc.store.Backup(ctx)
	if err != nil {
		return Gallery{}, err
	}
	applied := domain.AppliedTheme(a.CustomCSS())
	return Gallery{Themes: all, Applied: applied, CanRestore: applied != "" && hasBackup}, nil
}

func (uc *themes) Apply(ctx context.Context, id string) (AppearanceResponse, error) {
	t, err := uc.find(ctx, id)
	if err != nil {
		return AppearanceResponse{}, err
	}
	a, err := uc.repo.Load(ctx)
	if err != nil {
		return AppearanceResponse{}, fmt.Errorf("load appearance: %w", err)
	}
	// Own style is kept; the CSS of an earlier theme is not worth keeping.
	if !domain.IsThemeCSS(a.CustomCSS()) {
		if err := uc.store.SaveBackup(ctx, StyleBackup{Theme: a.Theme(), CSS: a.CustomCSS(), AccentColor: a.AccentColor()}); err != nil {
			return AppearanceResponse{}, err
		}
	}
	next, err := a.WithStyle(t.CSS(), t.Scheme, "")
	if err != nil {
		return AppearanceResponse{}, err
	}
	if err := uc.repo.Save(ctx, next); err != nil {
		return AppearanceResponse{}, fmt.Errorf("save appearance: %w", err)
	}
	return toResponse(next), nil
}

func (uc *themes) Restore(ctx context.Context) (AppearanceResponse, error) {
	b, ok, err := uc.store.Backup(ctx)
	if err != nil {
		return AppearanceResponse{}, err
	}
	a, err := uc.repo.Load(ctx)
	if err != nil {
		return AppearanceResponse{}, fmt.Errorf("load appearance: %w", err)
	}
	if !ok || !domain.IsThemeCSS(a.CustomCSS()) {
		return AppearanceResponse{}, kernel.NewPublicError(kernel.ErrConflict, "There is no custom CSS to restore.")
	}
	next, err := a.WithStyle(b.CSS, b.Theme, b.AccentColor)
	if err != nil {
		return AppearanceResponse{}, err
	}
	if err := uc.repo.Save(ctx, next); err != nil {
		return AppearanceResponse{}, fmt.Errorf("save appearance: %w", err)
	}
	return toResponse(next), nil
}

// maxImported bounds the imported themes.
const maxImported = 50

func (uc *themes) Import(ctx context.Context, data []byte) (domain.ThemeDef, error) {
	t, err := uc.codec.Decode(data)
	if err != nil {
		var ve *kernel.ValidationError
		if errors.As(err, &ve) {
			return domain.ThemeDef{}, err
		}
		return domain.ThemeDef{}, kernel.NewValidationError("file", "The file is not a trisse-home theme (JSON).")
	}
	imported, err := uc.store.Imported(ctx)
	if err != nil {
		return domain.ThemeDef{}, err
	}
	if len(imported) >= maxImported {
		return domain.ThemeDef{}, kernel.NewValidationError("file", "At most 50 imported themes are supported.")
	}
	// An id that is taken (by a built-in or another import) gets a suffix.
	base := t.ID
	for n := 2; ; n++ {
		if _, err := uc.find(ctx, t.ID); err != nil {
			break
		}
		t.ID = base + "-" + strconv.Itoa(n)
	}
	t.BuiltIn = false
	if err := uc.store.SaveImported(ctx, t); err != nil {
		return domain.ThemeDef{}, fmt.Errorf("save theme: %w", err)
	}
	return t, nil
}

func (uc *themes) Export(ctx context.Context, id string) ([]byte, error) {
	t, err := uc.find(ctx, id)
	if err != nil {
		return nil, err
	}
	return uc.codec.Encode(t)
}

func (uc *themes) Delete(ctx context.Context, id string) error {
	t, err := uc.find(ctx, id)
	if err != nil {
		return err
	}
	if t.BuiltIn {
		return kernel.NewValidationError("id", "Built-in themes cannot be deleted.")
	}
	return uc.store.DeleteImported(ctx, id)
}
