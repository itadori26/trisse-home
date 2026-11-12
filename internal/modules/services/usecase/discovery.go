package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/services/domain"
)

// ContainerSource lists the containers with their labels (output port,
// implemented in the composition root on top of the docker module).
type ContainerSource interface {
	Containers(ctx context.Context) ([]domain.Container, error)
}

// SyncResult is the outcome of the last sync.
type SyncResult struct {
	At                              time.Time
	Error                           string
	Created, Updated, Hidden, Shown int
	// Skipped counts containers whose labels are invalid (e.g. a bad URL).
	Skipped int
}

// DiscoveryStore persists the discovery settings and the last result.
type DiscoveryStore interface {
	DiscoverySettings(ctx context.Context) (domain.DiscoverySettings, error)
	SaveDiscoverySettings(ctx context.Context, s domain.DiscoverySettings) error
	LastSync(ctx context.Context) (SyncResult, error)
	SaveLastSync(ctx context.Context, r SyncResult) error
}

// DiscoveryInput carries the editable discovery settings.
type DiscoveryInput struct {
	Enabled         bool
	IntervalMinutes int
	Rule            string
}

// DiscoveryView is the discovery state shown in the settings.
type DiscoveryView struct {
	// Available is false when the docker module is not configured.
	Available bool
	Settings  domain.DiscoverySettings
	Last      SyncResult
}

// Discovery is the input port of Docker autodiscovery (#78).
type Discovery interface {
	Get(ctx context.Context) (DiscoveryView, error)
	Save(ctx context.Context, in DiscoveryInput) (DiscoveryView, error)
	// Sync runs a sync now, regardless of the interval.
	Sync(ctx context.Context) (DiscoveryView, error)
	// SyncIfDue is run by the background job.
	SyncIfDue(ctx context.Context) error
}

// errDockerUnreachable is shown to the user; details go to the log.
const errDockerUnreachable = "The Docker API is not reachable."

type discovery struct {
	d      Deps
	source ContainerSource
	store  DiscoveryStore
	// mu serialises syncs so the job and a manual sync never create a tile twice.
	mu sync.Mutex
}

// NewDiscovery returns the discovery use cases; source may be nil when the
// docker module is disabled.
func NewDiscovery(d Deps, source ContainerSource, store DiscoveryStore) Discovery {
	return &discovery{d: d, source: source, store: store}
}

// Get implements Discovery.
func (uc *discovery) Get(ctx context.Context) (DiscoveryView, error) {
	s, err := uc.store.DiscoverySettings(ctx)
	if err != nil {
		return DiscoveryView{}, fmt.Errorf("load discovery settings: %w", err)
	}
	last, err := uc.store.LastSync(ctx)
	if err != nil {
		return DiscoveryView{}, fmt.Errorf("load last sync: %w", err)
	}
	return DiscoveryView{Available: uc.source != nil, Settings: s, Last: last}, nil
}

// Save implements Discovery.
func (uc *discovery) Save(ctx context.Context, in DiscoveryInput) (DiscoveryView, error) {
	s := domain.DiscoverySettings{Enabled: in.Enabled, IntervalMinutes: in.IntervalMinutes, Rule: domain.MergeRule(in.Rule)}
	if err := s.Validate(); err != nil {
		return DiscoveryView{}, err
	}
	if s.Enabled && uc.source == nil {
		return DiscoveryView{}, domain.ErrDiscoveryUnavailable
	}
	if err := uc.store.SaveDiscoverySettings(ctx, s); err != nil {
		return DiscoveryView{}, fmt.Errorf("save discovery settings: %w", err)
	}
	return uc.Get(ctx)
}

// Sync implements Discovery. A Docker error is stored and shown in the
// view; it is not an error of the request.
func (uc *discovery) Sync(ctx context.Context) (DiscoveryView, error) {
	if uc.source == nil {
		return DiscoveryView{}, domain.ErrDiscoveryUnavailable
	}
	s, err := uc.store.DiscoverySettings(ctx)
	if err != nil {
		return DiscoveryView{}, fmt.Errorf("load discovery settings: %w", err)
	}
	if err := uc.sync(ctx, s.Rule); err != nil && !errors.Is(err, errSourceFailed) {
		return DiscoveryView{}, err
	}
	return uc.Get(ctx)
}

// SyncIfDue implements Discovery.
func (uc *discovery) SyncIfDue(ctx context.Context) error {
	if uc.source == nil {
		return nil
	}
	s, err := uc.store.DiscoverySettings(ctx)
	if err != nil {
		return fmt.Errorf("load discovery settings: %w", err)
	}
	last, err := uc.store.LastSync(ctx)
	if err != nil {
		return fmt.Errorf("load last sync: %w", err)
	}
	if !s.Due(last.At, uc.d.Clock.Now()) {
		return nil
	}
	return uc.sync(ctx, s.Rule)
}

var errSourceFailed = errors.New("list containers")

func (uc *discovery) sync(ctx context.Context, rule domain.MergeRule) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	res := SyncResult{At: uc.d.Clock.Now()}
	containers, err := uc.source.Containers(ctx)
	if err != nil {
		res.Error = errDockerUnreachable
		if saveErr := uc.store.SaveLastSync(ctx, res); saveErr != nil {
			return fmt.Errorf("save last sync: %w", saveErr)
		}
		return fmt.Errorf("%w: %w", errSourceFailed, err)
	}
	var found []domain.DiscoveredService
	for _, c := range containers {
		d, ok := domain.ParseLabels(c)
		if !ok {
			continue
		}
		if d.Validate() != nil {
			res.Skipped++
			continue
		}
		found = append(found, d)
	}
	existing, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list services: %w", err)
	}
	if err := uc.apply(ctx, domain.PlanSync(existing, found, rule), len(existing), &res); err != nil {
		return err
	}
	if err := uc.store.SaveLastSync(ctx, res); err != nil {
		return fmt.Errorf("save last sync: %w", err)
	}
	return nil
}

func (uc *discovery) apply(ctx context.Context, p domain.SyncPlan, position int, res *SyncResult) error {
	for _, d := range p.Create {
		s, err := domain.NewService(uc.d.IDs.NewID(), d.Name, d.URL, "", d.Icon, d.Description, false, position)
		if err != nil {
			return fmt.Errorf("new discovered service: %w", err) // validated before planning
		}
		s.Container = d.Container
		if err := uc.d.Repo.Create(ctx, s); err != nil {
			return fmt.Errorf("create discovered service: %w", err)
		}
		position++
	}
	// Show runs before Update: a tile can be in both lists, and the update
	// carries Hidden=false as well.
	for _, group := range [][]domain.Service{p.Show, p.Update, p.Hide} {
		for _, s := range group {
			if err := uc.d.Repo.Update(ctx, s); err != nil {
				return fmt.Errorf("update discovered service: %w", err)
			}
		}
	}
	for _, s := range p.Hide {
		uc.d.Status.Delete(s.ID)
	}
	res.Created, res.Updated, res.Hidden, res.Shown = len(p.Create), len(p.Update), len(p.Hide), len(p.Show)
	return nil
}
