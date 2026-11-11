package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/shared/kernel"
)

// maxParallelChecks bounds concurrent outbound health checks.
const maxParallelChecks = 8

// ServiceInput carries the editable fields of a service.
type ServiceInput struct {
	Name, URL, CheckURL, Icon, Description string
	SkipTLSVerify                          bool
}

// ServiceView is a service with its latest status.
type ServiceView struct {
	ID, Name, URL, CheckURL, Icon, Description string
	SkipTLSVerify                              bool
	Position                                   int
	State                                      string
	HTTPStatus                                 int
	LatencyMs                                  int64
	CheckedAt                                  time.Time
	Error                                      string
	CertExpiresAt                              time.Time
	CertExpiresSoon                            bool
}

// Services is the input port of the services module.
type Services interface {
	List(ctx context.Context) ([]ServiceView, error)
	Create(ctx context.Context, in ServiceInput) (ServiceView, error)
	Update(ctx context.Context, id string, in ServiceInput) (ServiceView, error)
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ids []string) error
	CheckAll(ctx context.Context) error
	// History returns the response time history per service ID.
	History(ctx context.Context, hours int) (map[string][]LatencyPointView, error)
}

// Repository persists services (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Service, error)
	Get(ctx context.Context, id string) (domain.Service, error)
	Create(ctx context.Context, s domain.Service) error
	Update(ctx context.Context, s domain.Service) error
	Delete(ctx context.Context, id string) error
}

// StatusStore caches the latest health check results (output port).
type StatusStore interface {
	Get(serviceID string) (domain.Status, bool)
	Put(s domain.Status)
	Delete(serviceID string)
}

// HealthChecker performs one HTTP health check (output port).
type HealthChecker interface {
	Check(ctx context.Context, url string, skipTLSVerify bool) domain.CheckResult
}

// HistoryStore keeps the health check history (output port).
type HistoryStore interface {
	Add(ctx context.Context, p domain.LatencyPoint) error
	Since(ctx context.Context, t time.Time) ([]domain.LatencyPoint, error)
	PruneBefore(ctx context.Context, t time.Time) error
}

// LatencyPointView is one point of a service's response time history.
type LatencyPointView struct {
	At        time.Time
	LatencyMs float64
	Online    bool
}

// Deps bundles the output ports.
type Deps struct {
	Repo    Repository
	Status  StatusStore
	Checker HealthChecker
	History HistoryStore
	// Observations receives the health of every check (e.g. for notifications).
	Observations kernel.ObservationSink
	Clock        kernel.Clock
	IDs          kernel.IDGenerator
}

type services struct{ d Deps }

// NewServices returns the services use cases.
func NewServices(d Deps) Services { return &services{d: d} }

// List implements Services.
func (uc *services) List(ctx context.Context) ([]ServiceView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	out := make([]ServiceView, 0, len(list))
	for _, s := range list {
		out = append(out, uc.view(s))
	}
	return out, nil
}

// Create implements Services; the new tile is checked immediately.
func (uc *services) Create(ctx context.Context, in ServiceInput) (ServiceView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return ServiceView{}, fmt.Errorf("list services: %w", err)
	}
	s, err := domain.NewService(uc.d.IDs.NewID(), in.Name, in.URL, in.CheckURL, in.Icon, in.Description, in.SkipTLSVerify, len(list))
	if err != nil {
		return ServiceView{}, err
	}
	if err := uc.d.Repo.Create(ctx, s); err != nil {
		return ServiceView{}, fmt.Errorf("create service: %w", err)
	}
	uc.check(ctx, s)
	return uc.view(s), nil
}

// Update implements Services.
func (uc *services) Update(ctx context.Context, id string, in ServiceInput) (ServiceView, error) {
	old, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return ServiceView{}, err
	}
	s, err := domain.NewService(id, in.Name, in.URL, in.CheckURL, in.Icon, in.Description, in.SkipTLSVerify, old.Position)
	if err != nil {
		return ServiceView{}, err
	}
	if err := uc.d.Repo.Update(ctx, s); err != nil {
		return ServiceView{}, fmt.Errorf("update service: %w", err)
	}
	uc.check(ctx, s)
	return uc.view(s), nil
}

// Delete implements Services.
func (uc *services) Delete(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	if err := uc.d.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	uc.d.Status.Delete(id)
	return nil
}

// Reorder implements Services; ids must list every service exactly once.
func (uc *services) Reorder(ctx context.Context, ids []string) error {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list services: %w", err)
	}
	byID := make(map[string]domain.Service, len(list))
	for _, s := range list {
		byID[s.ID] = s
	}
	if len(ids) != len(list) {
		return kernel.NewValidationError("ids", "The new order must contain every service exactly once.")
	}
	for i, id := range ids {
		s, ok := byID[id]
		if !ok {
			return kernel.NewValidationError("ids", "The new order must contain every service exactly once.")
		}
		delete(byID, id)
		s.Position = i
		if err := uc.d.Repo.Update(ctx, s); err != nil {
			return fmt.Errorf("update position: %w", err)
		}
	}
	return nil
}

// CheckAll implements Services; it is run by the background job.
func (uc *services) CheckAll(ctx context.Context) error {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list services: %w", err)
	}
	if err := uc.d.History.PruneBefore(ctx, uc.d.Clock.Now().Add(-domain.HistoryRetention)); err != nil {
		return fmt.Errorf("prune history: %w", err)
	}
	sem := make(chan struct{}, maxParallelChecks)
	var wg sync.WaitGroup
	for _, s := range list {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			uc.check(ctx, s)
		}()
	}
	wg.Wait()
	return nil
}

func (uc *services) check(ctx context.Context, s domain.Service) {
	now := uc.d.Clock.Now()
	st := domain.StatusFromCheck(s.ID, uc.d.Checker.Check(ctx, s.CheckURL, s.SkipTLSVerify), now)
	uc.d.Status.Put(st)
	uc.d.Observations.Observe(ctx, kernel.Observation{Kind: "service", Key: s.ID, Label: s.Name, Healthy: st.State == domain.StateOnline, Detail: st.Error})
	if !st.CertExpiresAt.IsZero() {
		days := st.CertExpiresAt.Sub(now).Hours() / 24
		uc.d.Observations.Observe(ctx, kernel.Observation{Kind: "tls", Key: s.ID, Label: s.Name, Healthy: true, Value: days,
			Detail: "The certificate expires on " + st.CertExpiresAt.Format("2006-01-02") + "."})
	}
	// History is best effort: a failed write must not hide the live status.
	_ = uc.d.History.Add(ctx, domain.LatencyPoint{ServiceID: s.ID, At: now, LatencyMs: float64(st.Latency.Milliseconds()), Online: st.State == domain.StateOnline})
}

// History implements Services.
func (uc *services) History(ctx context.Context, hours int) (map[string][]LatencyPointView, error) {
	maxHours := int(domain.HistoryRetention / time.Hour)
	if hours <= 0 || hours > maxHours {
		hours = maxHours
	}
	points, err := uc.d.History.Since(ctx, uc.d.Clock.Now().Add(-time.Duration(hours)*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	byService := map[string][]domain.LatencyPoint{}
	for _, p := range points {
		byService[p.ServiceID] = append(byService[p.ServiceID], p)
	}
	out := make(map[string][]LatencyPointView, len(byService))
	for id, ps := range byService {
		for _, p := range domain.DownsampleLatency(ps, domain.MaxChartPoints) {
			out[id] = append(out[id], LatencyPointView{At: p.At, LatencyMs: p.LatencyMs, Online: p.Online})
		}
	}
	return out, nil
}

func (uc *services) view(s domain.Service) ServiceView {
	v := ServiceView{
		ID: s.ID, Name: s.Name, URL: s.URL, CheckURL: s.CheckURL, Icon: s.Icon, Description: s.Description,
		SkipTLSVerify: s.SkipTLSVerify, Position: s.Position, State: string(domain.StateUnknown),
	}
	if st, ok := uc.d.Status.Get(s.ID); ok {
		v.State, v.HTTPStatus, v.LatencyMs, v.CheckedAt, v.Error = string(st.State), st.HTTPStatus, st.Latency.Milliseconds(), st.CheckedAt, st.Error
		v.CertExpiresAt, v.CertExpiresSoon = st.CertExpiresAt, st.CertificateExpiresWithin(domain.DefaultCertWarning, uc.d.Clock.Now())
	}
	return v
}
