package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/study/domain"
	"trisse-home/internal/shared/kernel"
)

// Study is the input port of the module.
type Study interface {
	Modules(ctx context.Context) ([]domain.Module, error)
	Create(ctx context.Context, in domain.ModuleInput) (domain.Module, error)
	Update(ctx context.Context, id string, in domain.ModuleInput) (domain.Module, error)
	Delete(ctx context.Context, id string) error
	TotalECTS(ctx context.Context) (float64, error)
	SetTotalECTS(ctx context.Context, v float64) (float64, error)
	// Progress sums up all modules against the total ECTS.
	Progress(ctx context.Context) (domain.Progress, []domain.Module, error)
}

// Repository stores modules and the total ECTS (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Module, error)
	Get(ctx context.Context, id string) (domain.Module, error)
	Save(ctx context.Context, m domain.Module) error
	Delete(ctx context.Context, id string) error
	// TotalECTS returns domain.DefaultTotalECTS while nothing is stored.
	TotalECTS(ctx context.Context) (float64, error)
	SetTotalECTS(ctx context.Context, v float64) error
}

type study struct {
	repo Repository
	ids  kernel.IDGenerator
}

// New returns the use cases.
func New(repo Repository, ids kernel.IDGenerator) Study { return &study{repo: repo, ids: ids} }

func (uc *study) Modules(ctx context.Context) ([]domain.Module, error) {
	ms, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	domain.Sort(ms)
	return ms, nil
}

func (uc *study) Create(ctx context.Context, in domain.ModuleInput) (domain.Module, error) {
	ms, err := uc.repo.List(ctx)
	if err != nil {
		return domain.Module{}, err
	}
	if len(ms) >= domain.MaxModules {
		return domain.Module{}, kernel.NewValidationError("name", "At most 200 modules are supported.")
	}
	m, err := domain.NewModule(uc.ids.NewID(), in)
	if err != nil {
		return domain.Module{}, err
	}
	if err := uc.repo.Save(ctx, m); err != nil {
		return domain.Module{}, fmt.Errorf("save module: %w", err)
	}
	return m, nil
}

func (uc *study) Update(ctx context.Context, id string, in domain.ModuleInput) (domain.Module, error) {
	if _, err := uc.repo.Get(ctx, id); err != nil {
		return domain.Module{}, err
	}
	m, err := domain.NewModule(id, in)
	if err != nil {
		return domain.Module{}, err
	}
	if err := uc.repo.Save(ctx, m); err != nil {
		return domain.Module{}, fmt.Errorf("save module: %w", err)
	}
	return m, nil
}

func (uc *study) Delete(ctx context.Context, id string) error { return uc.repo.Delete(ctx, id) }

func (uc *study) TotalECTS(ctx context.Context) (float64, error) { return uc.repo.TotalECTS(ctx) }

func (uc *study) SetTotalECTS(ctx context.Context, v float64) (float64, error) {
	v, err := domain.NewTotalECTS(v)
	if err != nil {
		return 0, err
	}
	if err := uc.repo.SetTotalECTS(ctx, v); err != nil {
		return 0, fmt.Errorf("save total: %w", err)
	}
	return v, nil
}

func (uc *study) Progress(ctx context.Context) (domain.Progress, []domain.Module, error) {
	ms, err := uc.Modules(ctx)
	if err != nil {
		return domain.Progress{}, nil, err
	}
	total, err := uc.repo.TotalECTS(ctx)
	if err != nil {
		return domain.Progress{}, nil, err
	}
	return domain.ComputeProgress(ms, total), ms, nil
}
