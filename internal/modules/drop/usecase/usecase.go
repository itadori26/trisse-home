package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"trisse-home/internal/modules/drop/domain"
	"trisse-home/internal/shared/kernel"
)

// Repository stores items and file contents (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Item, error)
	Get(ctx context.Context, id string) (domain.Item, error)
	// Add stores an item; data is the file content (nil for text).
	Add(ctx context.Context, i domain.Item, data []byte) error
	Delete(ctx context.Context, id string) error
	File(ctx context.Context, id string) ([]byte, error)
	// DeleteExpired removes items that expired before now and returns how many.
	DeleteExpired(ctx context.Context, now time.Time) (int, error)
}

// Drop is the input port of the module.
type Drop interface {
	AddText(ctx context.Context, text string, expiresIn time.Duration) (domain.Item, error)
	AddFile(ctx context.Context, name, contentType string, data []byte, expiresIn time.Duration) (domain.Item, error)
	// List returns the items that have not expired, newest first.
	List(ctx context.Context) ([]domain.Item, error)
	Delete(ctx context.Context, id string) error
	File(ctx context.Context, id string) (domain.Item, []byte, error)
	// PurgeExpired deletes expired items (background job).
	PurgeExpired(ctx context.Context) error
}

type drop struct {
	repo  Repository
	clock kernel.Clock
	ids   kernel.IDGenerator
}

// New returns the use cases.
func New(repo Repository, clock kernel.Clock, ids kernel.IDGenerator) Drop {
	return &drop{repo: repo, clock: clock, ids: ids}
}

func (uc *drop) add(ctx context.Context, i domain.Item, data []byte) (domain.Item, error) {
	items, err := uc.repo.List(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	if err := domain.CheckQuota(items, uc.clock.Now(), i.Size); err != nil {
		return domain.Item{}, err
	}
	if err := uc.repo.Add(ctx, i, data); err != nil {
		return domain.Item{}, fmt.Errorf("add item: %w", err)
	}
	return i, nil
}

func (uc *drop) AddText(ctx context.Context, text string, expiresIn time.Duration) (domain.Item, error) {
	i, err := domain.NewText(uc.ids.NewID(), text, expiresIn, uc.clock.Now())
	if err != nil {
		return domain.Item{}, err
	}
	return uc.add(ctx, i, nil)
}

func (uc *drop) AddFile(ctx context.Context, name, contentType string, data []byte, expiresIn time.Duration) (domain.Item, error) {
	i, err := domain.NewFile(uc.ids.NewID(), name, contentType, int64(len(data)), expiresIn, uc.clock.Now())
	if err != nil {
		return domain.Item{}, err
	}
	return uc.add(ctx, i, data)
}

func (uc *drop) List(ctx context.Context) ([]domain.Item, error) {
	items, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	now := uc.clock.Now()
	items = slices.DeleteFunc(items, func(i domain.Item) bool { return i.Expired(now) })
	slices.SortFunc(items, func(a, b domain.Item) int { return b.CreatedAt.Compare(a.CreatedAt) })
	return items, nil
}

func (uc *drop) get(ctx context.Context, id string) (domain.Item, error) {
	i, err := uc.repo.Get(ctx, id)
	if err != nil {
		return domain.Item{}, err
	}
	if i.Expired(uc.clock.Now()) {
		return domain.Item{}, domain.ErrNotFound
	}
	return i, nil
}

func (uc *drop) Delete(ctx context.Context, id string) error {
	if _, err := uc.repo.Get(ctx, id); err != nil {
		return err
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *drop) File(ctx context.Context, id string) (domain.Item, []byte, error) {
	i, err := uc.get(ctx, id)
	if err != nil {
		return domain.Item{}, nil, err
	}
	if i.Kind != domain.KindFile {
		return domain.Item{}, nil, domain.ErrNotFound
	}
	data, err := uc.repo.File(ctx, id)
	if err != nil {
		return domain.Item{}, nil, err
	}
	return i, data, nil
}

func (uc *drop) PurgeExpired(ctx context.Context) error {
	_, err := uc.repo.DeleteExpired(ctx, uc.clock.Now())
	return err
}
