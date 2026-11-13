package usecase

import (
	"context"
	"errors"
	"fmt"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/shared/kernel"
)

// ErrShortcutNotFound is returned when deleting an unknown shortcut.
var ErrShortcutNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The shortcut does not exist.")

// ErrShortcutExists is returned when a keyword is already taken.
var ErrShortcutExists = kernel.NewPublicError(kernel.ErrConflict, "A shortcut with this keyword already exists.")

// ProviderView describes a selectable provider.
type ProviderView struct {
	ID   string
	Name string
}

// ShortcutView describes a custom shortcut.
type ShortcutView struct {
	Keyword     string
	Name        string
	URLTemplate string
}

// BarView describes the options of the search field (#221).
type BarView struct {
	OpenInNewTab           bool
	TodoPrefix, NotePrefix string
	BookmarkPrefix         string
}

// SettingsResponse describes the search configuration.
type SettingsResponse struct {
	DefaultProvider string
	Providers       []ProviderView
	Shortcuts       []ShortcutView
	Bar             BarView
}

// SetBarRequest is the input of SetBarOptions.
type SetBarRequest struct {
	OpenInNewTab                           bool
	TodoPrefix, NotePrefix, BookmarkPrefix string
}

// AddShortcutRequest is the input of AddShortcut.
type AddShortcutRequest struct {
	Keyword     string
	Name        string
	URLTemplate string
}

// Searcher is the input port of the search module.
type Searcher interface {
	Resolve(ctx context.Context, query string) (string, error)
	Settings(ctx context.Context) (SettingsResponse, error)
	SetDefaultProvider(ctx context.Context, providerID string) (SettingsResponse, error)
	SetBarOptions(ctx context.Context, req SetBarRequest) (SettingsResponse, error)
	AddShortcut(ctx context.Context, req AddShortcutRequest) (SettingsResponse, error)
	DeleteShortcut(ctx context.Context, keyword string) (SettingsResponse, error)
}

// Repository persists the search settings (output port).
type Repository interface {
	// DefaultProvider returns "" if none has been chosen yet.
	DefaultProvider(ctx context.Context) (string, error)
	SetDefaultProvider(ctx context.Context, id string) error
	// BarOptions returns the defaults if none have been saved yet.
	BarOptions(ctx context.Context) (domain.BarOptions, error)
	SetBarOptions(ctx context.Context, o domain.BarOptions) error
	Shortcuts(ctx context.Context) ([]domain.Shortcut, error)
	// AddShortcut returns ErrShortcutExists for duplicate keywords.
	AddShortcut(ctx context.Context, s domain.Shortcut) error
	// DeleteShortcut returns ErrShortcutNotFound for unknown keywords.
	DeleteShortcut(ctx context.Context, keyword string) error
}

type searcher struct{ repo Repository }

// NewSearcher returns the Searcher use cases.
func NewSearcher(repo Repository) Searcher { return &searcher{repo: repo} }

func (uc *searcher) provider(ctx context.Context) (domain.Provider, error) {
	id, err := uc.repo.DefaultProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("load default provider: %w", err)
	}
	if id == "" {
		id = domain.DefaultProviderID
	}
	p, err := domain.ProviderByID(id)
	if err != nil {
		// A provider removed in a newer version falls back to the default.
		return domain.ProviderByID(domain.DefaultProviderID)
	}
	return p, nil
}

// Resolve implements Searcher.
func (uc *searcher) Resolve(ctx context.Context, query string) (string, error) {
	p, err := uc.provider(ctx)
	if err != nil {
		return "", err
	}
	shortcuts, err := uc.repo.Shortcuts(ctx)
	if err != nil {
		return "", fmt.Errorf("load shortcuts: %w", err)
	}
	return domain.Resolve(query, p, shortcuts)
}

// Settings implements Searcher.
func (uc *searcher) Settings(ctx context.Context) (SettingsResponse, error) {
	p, err := uc.provider(ctx)
	if err != nil {
		return SettingsResponse{}, err
	}
	shortcuts, err := uc.repo.Shortcuts(ctx)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("load shortcuts: %w", err)
	}
	bar, err := uc.repo.BarOptions(ctx)
	if err != nil {
		return SettingsResponse{}, fmt.Errorf("load search field options: %w", err)
	}
	pre := bar.Prefixes()
	res := SettingsResponse{DefaultProvider: p.ID(), Providers: []ProviderView{}, Shortcuts: []ShortcutView{},
		Bar: BarView{OpenInNewTab: bar.OpenInNewTab(), TodoPrefix: pre.Todo, NotePrefix: pre.Note, BookmarkPrefix: pre.Bookmark}}
	for _, bp := range domain.Providers() {
		res.Providers = append(res.Providers, ProviderView{ID: bp.ID(), Name: bp.Name()})
	}
	for _, s := range shortcuts {
		res.Shortcuts = append(res.Shortcuts, ShortcutView{Keyword: s.Keyword(), Name: s.Name(), URLTemplate: s.Template()})
	}
	return res, nil
}

// SetDefaultProvider implements Searcher.
func (uc *searcher) SetDefaultProvider(ctx context.Context, providerID string) (SettingsResponse, error) {
	if _, err := domain.ProviderByID(providerID); err != nil {
		return SettingsResponse{}, err
	}
	if err := uc.repo.SetDefaultProvider(ctx, providerID); err != nil {
		return SettingsResponse{}, fmt.Errorf("save default provider: %w", err)
	}
	return uc.Settings(ctx)
}

// SetBarOptions implements Searcher.
func (uc *searcher) SetBarOptions(ctx context.Context, req SetBarRequest) (SettingsResponse, error) {
	o, err := domain.NewBarOptions(req.OpenInNewTab, domain.CapturePrefixes{Todo: req.TodoPrefix, Note: req.NotePrefix, Bookmark: req.BookmarkPrefix})
	if err != nil {
		return SettingsResponse{}, err
	}
	if err := uc.repo.SetBarOptions(ctx, o); err != nil {
		return SettingsResponse{}, fmt.Errorf("save search field options: %w", err)
	}
	return uc.Settings(ctx)
}

// AddShortcut implements Searcher.
func (uc *searcher) AddShortcut(ctx context.Context, req AddShortcutRequest) (SettingsResponse, error) {
	s, err := domain.NewShortcut(req.Keyword, req.Name, req.URLTemplate)
	if err != nil {
		return SettingsResponse{}, err
	}
	if err := uc.repo.AddShortcut(ctx, s); err != nil {
		if errors.Is(err, ErrShortcutExists) {
			return SettingsResponse{}, err
		}
		return SettingsResponse{}, fmt.Errorf("add shortcut: %w", err)
	}
	return uc.Settings(ctx)
}

// DeleteShortcut implements Searcher.
func (uc *searcher) DeleteShortcut(ctx context.Context, keyword string) (SettingsResponse, error) {
	if err := uc.repo.DeleteShortcut(ctx, keyword); err != nil {
		return SettingsResponse{}, err
	}
	return uc.Settings(ctx)
}
