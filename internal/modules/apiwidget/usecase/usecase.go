package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/shared/kernel"
)

// Values is the result of one fetch of an endpoint.
type Values struct {
	Name      string
	Values    []domain.Value
	FetchedAt time.Time
	// Error is set when the request failed (a translatable message); Values
	// are then empty.
	Error *kernel.PublicError
}

// Preview is the result of a test request from the settings.
type Preview struct {
	Values
	// Body is the beginning of the response, for building paths.
	Body string
}

// APIWidget is the input port of the module.
type APIWidget interface {
	Endpoints(ctx context.Context) ([]domain.Endpoint, error)
	CreateEndpoint(ctx context.Context, in domain.EndpointInput) (domain.Endpoint, error)
	UpdateEndpoint(ctx context.Context, id string, in domain.EndpointInput) (domain.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error
	// Preview runs the request of a draft (id = "" for a new endpoint) without storing it.
	Preview(ctx context.Context, id string, in domain.EndpointInput) (Preview, error)
	// FetchValues returns the values of an endpoint, at most one request per interval.
	FetchValues(ctx context.Context, id string) (Values, error)
}

// Response is a fetched JSON document.
type Response struct {
	Doc any
	// Raw is the beginning of the body (at most previewBytes).
	Raw []byte
}

// JSONFetcher requests an endpoint and decodes the JSON body (output port).
type JSONFetcher interface {
	Fetch(ctx context.Context, e domain.Endpoint) (Response, error)
}

// PathEvaluator validates and evaluates JSONPath expressions (output port).
type PathEvaluator interface {
	Validate(path string) error
	Evaluate(doc any, path string) (any, error)
}

// EndpointRepository stores endpoints with their encrypted header secret (output port).
type EndpointRepository interface {
	List(ctx context.Context) ([]domain.Endpoint, error)
	// Get returns domain.ErrNotFound for an unknown id.
	Get(ctx context.Context, id string) (domain.Endpoint, error)
	Save(ctx context.Context, e domain.Endpoint) error
	Delete(ctx context.Context, id string) error
}

// errorRetry bounds how long a failed request is cached, so a broken
// endpoint is not hammered and a fixed one recovers soon.
const errorRetry = time.Minute

type cached struct {
	values Values
	until  time.Time
}

type apiWidget struct {
	repo    EndpointRepository
	fetcher JSONFetcher
	paths   PathEvaluator
	clock   kernel.Clock
	ids     kernel.IDGenerator

	mu    sync.Mutex
	cache map[string]cached
}

// New returns the use cases.
func New(repo EndpointRepository, fetcher JSONFetcher, paths PathEvaluator, clock kernel.Clock, ids kernel.IDGenerator) APIWidget {
	return &apiWidget{repo: repo, fetcher: fetcher, paths: paths, clock: clock, ids: ids, cache: map[string]cached{}}
}

func (uc *apiWidget) Endpoints(ctx context.Context) ([]domain.Endpoint, error) {
	return uc.repo.List(ctx)
}

func (uc *apiWidget) CreateEndpoint(ctx context.Context, in domain.EndpointInput) (domain.Endpoint, error) {
	all, err := uc.repo.List(ctx)
	if err != nil {
		return domain.Endpoint{}, err
	}
	if len(all) >= domain.MaxEndpoints {
		return domain.Endpoint{}, kernel.NewValidationError("name", "At most 50 endpoints are supported.")
	}
	e, err := domain.NewEndpoint(uc.ids.NewID(), in, nil, uc.paths.Validate)
	if err != nil {
		return domain.Endpoint{}, err
	}
	if err := uc.repo.Save(ctx, e); err != nil {
		return domain.Endpoint{}, fmt.Errorf("save endpoint: %w", err)
	}
	return e, nil
}

func (uc *apiWidget) UpdateEndpoint(ctx context.Context, id string, in domain.EndpointInput) (domain.Endpoint, error) {
	stored, err := uc.repo.Get(ctx, id)
	if err != nil {
		return domain.Endpoint{}, err
	}
	e, err := domain.NewEndpoint(id, in, &stored, uc.paths.Validate)
	if err != nil {
		return domain.Endpoint{}, err
	}
	if err := uc.repo.Save(ctx, e); err != nil {
		return domain.Endpoint{}, fmt.Errorf("save endpoint: %w", err)
	}
	uc.forget(id)
	return e, nil
}

func (uc *apiWidget) DeleteEndpoint(ctx context.Context, id string) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	uc.forget(id)
	return nil
}

func (uc *apiWidget) Preview(ctx context.Context, id string, in domain.EndpointInput) (Preview, error) {
	var stored *domain.Endpoint
	if id != "" {
		s, err := uc.repo.Get(ctx, id)
		if err != nil {
			return Preview{}, err
		}
		stored = &s
	}
	e, err := domain.NewEndpoint(id, in, stored, uc.paths.Validate)
	if err != nil {
		return Preview{}, err
	}
	res, values := uc.fetch(ctx, e)
	return Preview{Values: values, Body: string(res.Raw)}, nil
}

func (uc *apiWidget) FetchValues(ctx context.Context, id string) (Values, error) {
	now := uc.clock.Now()
	uc.mu.Lock()
	c, ok := uc.cache[id]
	uc.mu.Unlock()
	if ok && now.Before(c.until) {
		return c.values, nil
	}
	e, err := uc.repo.Get(ctx, id)
	if err != nil {
		return Values{}, err
	}
	_, values := uc.fetch(ctx, e)
	until := now.Add(e.Interval)
	if values.Error != nil {
		until = now.Add(min(e.Interval, errorRetry))
	}
	uc.mu.Lock()
	uc.cache[id] = cached{values: values, until: until}
	uc.mu.Unlock()
	return values, nil
}

// fetch requests e and evaluates its extractions; request errors end up in
// Values.Error (shown in the widget), not as an error of the call.
func (uc *apiWidget) fetch(ctx context.Context, e domain.Endpoint) (Response, Values) {
	out := Values{Name: e.Name, FetchedAt: uc.clock.Now(), Values: []domain.Value{}}
	res, err := uc.fetcher.Fetch(ctx, e)
	if err != nil {
		if !errors.As(err, &out.Error) {
			out.Error = kernel.TargetFailure(kernel.ErrInvalid, err)
		}
		return res, out
	}
	for _, x := range e.Extractions {
		v, err := uc.paths.Evaluate(res.Doc, x.Path)
		if err != nil {
			out.Values = append(out.Values, x.Missing(err.Error()))
			continue
		}
		out.Values = append(out.Values, x.Evaluate(v))
	}
	return res, out
}

func (uc *apiWidget) forget(id string) {
	uc.mu.Lock()
	delete(uc.cache, id)
	uc.mu.Unlock()
}
