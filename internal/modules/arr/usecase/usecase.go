package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/arr/domain"
	"trisse-home/internal/shared/kernel"
)

// overviewTTL bounds how often the apps are asked; the overview is fetched
// on demand only, so nothing is polled while no widget is open.
const overviewTTL = time.Minute

// InstanceView is an instance without its API key.
type InstanceView struct {
	ID, App, Name, URL string
}

// InstanceInput carries the editable fields; an empty API key keeps the
// stored one when updating.
type InstanceInput struct {
	App, Name, URL, APIKey string
}

// Arr is the input port of the module (#86).
type Arr interface {
	Instances(ctx context.Context) ([]InstanceView, error)
	AddInstance(ctx context.Context, in InstanceInput) (InstanceView, error)
	UpdateInstance(ctx context.Context, id string, in InstanceInput) (InstanceView, error)
	DeleteInstance(ctx context.Context, id string) error
	Overview(ctx context.Context) (domain.Overview, error)
}

// InstanceRepository persists instances (output port).
type InstanceRepository interface {
	List(ctx context.Context) ([]domain.Instance, error)
	Get(ctx context.Context, id string) (domain.Instance, error)
	Save(ctx context.Context, in domain.Instance) error
	Delete(ctx context.Context, id string) error
}

// Provider reads one app (output port); one adapter per API version.
type Provider interface {
	// Check verifies URL and API key.
	Check(ctx context.Context, in domain.Instance) error
	Snapshot(ctx context.Context, in domain.Instance, from, to time.Time) (domain.Snapshot, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo      InstanceRepository
	Providers map[domain.App]Provider
	Clock     kernel.Clock
	IDs       kernel.IDGenerator
}

type arr struct {
	d       Deps
	mu      sync.Mutex
	cached  *domain.Overview
	fetched time.Time
}

// NewArr returns the use cases.
func NewArr(d Deps) Arr { return &arr{d: d} }

// Instances implements Arr.
func (uc *arr) Instances(ctx context.Context) ([]InstanceView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	out := make([]InstanceView, 0, len(list))
	for _, in := range list {
		out = append(out, view(in))
	}
	return out, nil
}

// AddInstance implements Arr; URL and API key are checked first.
func (uc *arr) AddInstance(ctx context.Context, in InstanceInput) (InstanceView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return InstanceView{}, fmt.Errorf("list instances: %w", err)
	}
	if len(list) >= domain.MaxInstances {
		return InstanceView{}, domain.ErrTooManyInstances
	}
	inst, err := domain.NewInstance(uc.d.IDs.NewID(), in.App, in.Name, in.URL, in.APIKey)
	if err != nil {
		return InstanceView{}, err
	}
	return uc.save(ctx, inst)
}

// UpdateInstance implements Arr.
func (uc *arr) UpdateInstance(ctx context.Context, id string, in InstanceInput) (InstanceView, error) {
	old, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return InstanceView{}, err
	}
	key := in.APIKey
	if key == "" {
		key = old.APIKey
	}
	inst, err := domain.NewInstance(id, in.App, in.Name, in.URL, key)
	if err != nil {
		return InstanceView{}, err
	}
	return uc.save(ctx, inst)
}

// DeleteInstance implements Arr.
func (uc *arr) DeleteInstance(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	if err := uc.d.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}
	uc.invalidate()
	return nil
}

func (uc *arr) save(ctx context.Context, in domain.Instance) (InstanceView, error) {
	if err := uc.d.Providers[in.App].Check(ctx, in); err != nil {
		return InstanceView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not connect to {app}: {detail}", "app", in.App.Title(), "detail", err.Error())
	}
	if err := uc.d.Repo.Save(ctx, in); err != nil {
		return InstanceView{}, fmt.Errorf("save instance: %w", err)
	}
	uc.invalidate()
	return view(in), nil
}

// Overview implements Arr. Instances are queried in parallel; one that
// fails is listed in Errors and does not hide the others.
func (uc *arr) Overview(ctx context.Context) (domain.Overview, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return domain.Overview{}, fmt.Errorf("list instances: %w", err)
	}
	if len(list) == 0 {
		return domain.Overview{}, domain.ErrNotConfigured
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.d.Clock.Now()
	if uc.cached != nil && now.Sub(uc.fetched) < overviewTTL {
		return *uc.cached, nil
	}
	from, to := now.Add(-12*time.Hour), now.AddDate(0, 0, domain.UpcomingDays)
	snaps := make([]domain.Snapshot, len(list))
	errs := make([]error, len(list))
	var wg sync.WaitGroup
	for i, in := range list {
		wg.Go(func() { snaps[i], errs[i] = uc.d.Providers[in.App].Snapshot(ctx, in, from, to) })
	}
	wg.Wait()
	names := make([]string, len(list))
	var failed []string
	for i, in := range list {
		names[i] = in.Name
		if errs[i] != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", in.Name, errs[i]))
		}
	}
	o := domain.Merge(names, snaps)
	o.Errors = failed
	uc.cached, uc.fetched = &o, now
	return o, nil
}

func (uc *arr) invalidate() {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = nil
}

func view(in domain.Instance) InstanceView {
	return InstanceView{ID: in.ID, App: string(in.App), Name: in.Name, URL: in.URL}
}
