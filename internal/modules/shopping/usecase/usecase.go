package usecase

import (
	"context"
	"errors"
	"fmt"

	"trisse-home/internal/modules/shopping/domain"
	"trisse-home/internal/shared/kernel"
)

// SettingsView never contains the password.
type SettingsView struct {
	Configured  bool
	ListURL     string
	Username    string
	HasPassword bool
}

// Shopping is the input port of the module.
type Shopping interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, listURL, username, password string) (SettingsView, error)
	List(ctx context.Context) ([]domain.Item, error)
	Add(ctx context.Context, summary string) (domain.Item, error)
	SetDone(ctx context.Context, id string, done bool) error
	Delete(ctx context.Context, id string) error
	// ClearCompleted deletes every completed item and returns how many.
	ClearCompleted(ctx context.Context) (int, error)
}

// SettingsRepository persists the settings (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Store reads and writes the items of a CalDAV task list (output port).
// Writes must fail with domain.ErrConflict when the item changed meanwhile.
type Store interface {
	List(ctx context.Context, s domain.Settings) ([]domain.Item, error)
	Add(ctx context.Context, s domain.Settings, uid, summary string) (domain.Item, error)
	SetDone(ctx context.Context, s domain.Settings, id string, done bool) error
	Delete(ctx context.Context, s domain.Settings, id string) error
}

type shopping struct {
	repo  SettingsRepository
	store Store
	ids   kernel.IDGenerator
}

// NewShopping returns the shopping list use cases.
func NewShopping(repo SettingsRepository, store Store, ids kernel.IDGenerator) Shopping {
	return &shopping{repo: repo, store: store, ids: ids}
}

// unavailable wraps store errors that are not domain errors.
func unavailable(err error) error {
	var pe *kernel.PublicError
	if errors.As(err, &pe) {
		return err
	}
	return kernel.TargetFailure(kernel.ErrUnavailable, err)
}

// Settings implements Shopping.
func (uc *shopping) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, ListURL: s.ListURL, Username: s.Username, HasPassword: true}, nil
}

// SaveSettings implements Shopping; the list is read once to test access.
func (uc *shopping) SaveSettings(ctx context.Context, listURL, username, password string) (SettingsView, error) {
	if password == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			if kernel.TargetChanged([]string{old.ListURL, old.Username}, []string{listURL, username}) {
				return SettingsView{}, kernel.SecretRequired("password")
			}
			password = old.Password
		}
	}
	s, err := domain.NewSettings(listURL, username, password)
	if err != nil {
		return SettingsView{}, err
	}
	if _, err := uc.store.List(ctx, s); err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	return SettingsView{Configured: true, ListURL: s.ListURL, Username: s.Username, HasPassword: true}, nil
}

// List implements Shopping.
func (uc *shopping) List(ctx context.Context) ([]domain.Item, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return nil, err
	}
	items, err := uc.store.List(ctx, s)
	if err != nil {
		return nil, unavailable(err)
	}
	return domain.ForDisplay(items), nil
}

// Add implements Shopping.
func (uc *shopping) Add(ctx context.Context, summary string) (domain.Item, error) {
	summary, err := domain.NormaliseSummary(summary)
	if err != nil {
		return domain.Item{}, err
	}
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	it, err := uc.store.Add(ctx, s, uc.ids.NewID()+"@trisse-home", summary)
	if err != nil {
		return domain.Item{}, unavailable(err)
	}
	return it, nil
}

// SetDone implements Shopping.
func (uc *shopping) SetDone(ctx context.Context, id string, done bool) error {
	if !domain.ValidItemID(id) {
		return domain.ErrItemNotFound
	}
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return err
	}
	if err := uc.store.SetDone(ctx, s, id, done); err != nil {
		return unavailable(err)
	}
	return nil
}

// Delete implements Shopping.
func (uc *shopping) Delete(ctx context.Context, id string) error {
	if !domain.ValidItemID(id) {
		return domain.ErrItemNotFound
	}
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return err
	}
	if err := uc.store.Delete(ctx, s, id); err != nil {
		return unavailable(err)
	}
	return nil
}

// ClearCompleted implements Shopping.
func (uc *shopping) ClearCompleted(ctx context.Context) (int, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return 0, err
	}
	items, err := uc.store.List(ctx, s)
	if err != nil {
		return 0, unavailable(err)
	}
	n := 0
	for _, it := range items {
		if !it.Done {
			continue
		}
		if err := uc.store.Delete(ctx, s, it.ID); err != nil && !errors.Is(err, domain.ErrItemNotFound) {
			return n, unavailable(err)
		}
		n++
	}
	return n, nil
}
