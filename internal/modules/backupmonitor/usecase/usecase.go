package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/backupmonitor/domain"
	"trisse-home/internal/shared/kernel"
)

// SourceInput carries the editable fields of a source.
type SourceInput struct {
	Name     string
	Schedule domain.Schedule
}

// PingEvent is what a ping reports.
type PingEvent string

// Ping events: a finished run, a failed run or the start of a run.
const (
	PingSuccess PingEvent = "success"
	PingFail    PingEvent = "fail"
	PingStart   PingEvent = "start"
)

// SourceView is a source with its evaluated state.
type SourceView struct {
	ID            string
	Name          string
	MaxAge        time.Duration
	Kind          string
	Cron          string
	Grace         time.Duration
	LastStartAt   time.Time
	LastDuration  time.Duration
	State         string
	LastSuccessAt time.Time
	LastFailureAt time.Time
	LastMessage   string
}

// CreatedSource is returned once after creating a source or rotating its
// token; the raw token is never stored.
type CreatedSource struct {
	Source SourceView
	Token  string
}

// Monitor is the input port of the module; every method is one use case.
type Monitor interface {
	List(ctx context.Context) ([]SourceView, error)
	Create(ctx context.Context, in SourceInput) (CreatedSource, error)
	Update(ctx context.Context, id string, in SourceInput) (SourceView, error)
	Delete(ctx context.Context, id string) error
	RotateToken(ctx context.Context, id string) (CreatedSource, error)
	// Ping records a run reported by a backup script.
	Ping(ctx context.Context, token string, event PingEvent, message string) error
	// Report sends the state of every source to the observation sink (job).
	Report(ctx context.Context) error
}

// Repository persists sources (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Source, error)
	Get(ctx context.Context, id string) (domain.Source, error)
	GetByToken(ctx context.Context, h domain.TokenHash) (domain.Source, error)
	Save(ctx context.Context, s domain.Source) error
	Delete(ctx context.Context, id string) error
}

// TokenGenerator creates unguessable ping tokens (output port).
type TokenGenerator interface {
	NewToken() (string, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo         Repository
	Tokens       TokenGenerator
	Clock        kernel.Clock
	IDs          kernel.IDGenerator
	Observations kernel.ObservationSink
}

type monitor struct{ d Deps }

// NewMonitor returns the backup monitor use cases.
func NewMonitor(d Deps) Monitor { return &monitor{d: d} }

func (uc *monitor) view(s domain.Source) SourceView {
	sch := s.Schedule()
	return SourceView{ID: s.ID, Name: s.Name, MaxAge: s.MaxAge, Kind: sch.Kind, Cron: sch.Cron, Grace: s.Grace,
		LastStartAt: s.LastStartAt, LastDuration: s.LastDuration, State: string(s.StateAt(uc.d.Clock.Now())),
		LastSuccessAt: s.LastSuccessAt, LastFailureAt: s.LastFailureAt, LastMessage: s.LastMessage}
}

// List implements Monitor.
func (uc *monitor) List(ctx context.Context) ([]SourceView, error) {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	out := make([]SourceView, 0, len(list))
	for _, s := range list {
		out = append(out, uc.view(s))
	}
	return out, nil
}

// Create implements Monitor.
func (uc *monitor) Create(ctx context.Context, in SourceInput) (CreatedSource, error) {
	token, err := uc.d.Tokens.NewToken()
	if err != nil {
		return CreatedSource{}, fmt.Errorf("generate token: %w", err)
	}
	s, err := domain.NewSource(uc.d.IDs.NewID(), in.Name, in.Schedule, domain.HashToken(token), uc.d.Clock.Now())
	if err != nil {
		return CreatedSource{}, err
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return CreatedSource{}, fmt.Errorf("save source: %w", err)
	}
	return CreatedSource{Source: uc.view(s), Token: token}, nil
}

// Update implements Monitor.
func (uc *monitor) Update(ctx context.Context, id string, in SourceInput) (SourceView, error) {
	s, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return SourceView{}, err
	}
	if err := s.Edit(in.Name, in.Schedule); err != nil {
		return SourceView{}, err
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SourceView{}, fmt.Errorf("save source: %w", err)
	}
	return uc.view(s), nil
}

// Delete implements Monitor.
func (uc *monitor) Delete(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	return uc.d.Repo.Delete(ctx, id)
}

// RotateToken implements Monitor; the old ping URL stops working.
func (uc *monitor) RotateToken(ctx context.Context, id string) (CreatedSource, error) {
	s, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return CreatedSource{}, err
	}
	token, err := uc.d.Tokens.NewToken()
	if err != nil {
		return CreatedSource{}, fmt.Errorf("generate token: %w", err)
	}
	s.TokenHash = domain.HashToken(token)
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return CreatedSource{}, fmt.Errorf("save source: %w", err)
	}
	return CreatedSource{Source: uc.view(s), Token: token}, nil
}

// Ping implements Monitor.
func (uc *monitor) Ping(ctx context.Context, token string, event PingEvent, message string) error {
	if token == "" {
		return domain.ErrSourceNotFound
	}
	s, err := uc.d.Repo.GetByToken(ctx, domain.HashToken(token))
	if err != nil {
		return err
	}
	switch event {
	case PingStart:
		s.RecordStart(uc.d.Clock.Now())
	case PingFail:
		s.RecordRun(false, message, uc.d.Clock.Now())
	default:
		s.RecordRun(true, message, uc.d.Clock.Now())
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return fmt.Errorf("save source: %w", err)
	}
	uc.observe(ctx, s)
	return nil
}

// Report implements Monitor.
func (uc *monitor) Report(ctx context.Context) error {
	list, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list sources: %w", err)
	}
	for _, s := range list {
		uc.observe(ctx, s)
	}
	return nil
}

func (uc *monitor) observe(ctx context.Context, s domain.Source) {
	state := s.StateAt(uc.d.Clock.Now())
	detail := string(state)
	if s.LastMessage != "" {
		detail += ": " + s.LastMessage
	}
	uc.d.Observations.Observe(ctx, kernel.Observation{
		Kind: s.Kind, Key: s.ID, Label: s.Name, Detail: detail,
		Healthy: state == domain.StateOK || state == domain.StateWaiting || state == domain.StateRunning,
	})
}
