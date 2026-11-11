package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/shared/kernel"
)

// deliveryTimeout bounds one push delivery.
const deliveryTimeout = 15 * time.Second

// SettingsInput is the input of SaveSettings; an empty token keeps the stored one.
type SettingsInput struct {
	Enabled   bool
	ServerURL string
	Topic     string
	Token     string
	Rules     domain.Rules
}

// SettingsView never contains the token.
type SettingsView struct {
	Enabled       bool
	ServerURL     string
	Topic         string
	HasToken      bool
	Rules         domain.Rules
	LastError     string
	LastDelivered time.Time
}

// Notifier is the input port of the module. It also implements
// kernel.ObservationSink so other modules can report to it.
type Notifier interface {
	kernel.ObservationSink
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	SendTest(ctx context.Context) error
}

// Repository persists the settings (output port); Load returns the defaults
// when nothing is stored.
type Repository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Sender delivers a message (output port).
type Sender interface {
	Send(ctx context.Context, s domain.Settings, m domain.Message) error
}

type notifier struct {
	repo   Repository
	sender Sender
	clock  kernel.Clock

	mu            sync.Mutex
	settings      *domain.Settings
	tracks        map[string]domain.Track
	lastError     string
	lastDelivered time.Time
	wg            sync.WaitGroup
}

// NewNotifier returns the notification use cases.
func NewNotifier(repo Repository, sender Sender, clock kernel.Clock) Notifier {
	return &notifier{repo: repo, sender: sender, clock: clock, tracks: map[string]domain.Track{}}
}

// current returns the cached settings, loading them once.
func (uc *notifier) current(ctx context.Context) (domain.Settings, error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.settings != nil {
		return *uc.settings, nil
	}
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("load notification settings: %w", err)
	}
	uc.settings = &s
	return s, nil
}

// Observe implements kernel.ObservationSink. Observations are always tracked
// (so enabling notifications later does not report old problems) but only
// delivered when notifications are enabled and the rule applies.
func (uc *notifier) Observe(ctx context.Context, o kernel.Observation) {
	s, err := uc.current(ctx)
	if err != nil {
		return
	}
	relevant, healthy := s.Rules.Judge(o)
	if !relevant {
		return
	}
	key := o.Kind + ":" + o.Key
	uc.mu.Lock()
	track, changed := uc.tracks[key].Advance(healthy)
	uc.tracks[key] = track
	uc.mu.Unlock()
	if !changed || !s.Enabled {
		return
	}
	msg := domain.MessageFor(o, healthy)
	uc.wg.Add(1)
	// Delivery runs in the background so slow push servers never delay the
	// health checks that produce observations.
	go func() {
		defer uc.wg.Done()
		dctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), deliveryTimeout)
		defer cancel()
		uc.record(uc.sender.Send(dctx, s, msg))
	}()
}

func (uc *notifier) record(err error) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if err != nil {
		uc.lastError = err.Error()
		return
	}
	uc.lastError, uc.lastDelivered = "", uc.clock.Now()
}

// Wait blocks until pending deliveries finished (used by tests).
func (uc *notifier) Wait() { uc.wg.Wait() }

// Settings implements Notifier.
func (uc *notifier) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.current(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	return uc.view(s), nil
}

func (uc *notifier) view(s domain.Settings) SettingsView {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return SettingsView{Enabled: s.Enabled, ServerURL: s.ServerURL, Topic: s.Topic, HasToken: s.Token != "",
		Rules: s.Rules, LastError: uc.lastError, LastDelivered: uc.lastDelivered}
}

// SaveSettings implements Notifier.
func (uc *notifier) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	old, err := uc.current(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	token := in.Token
	if token == "" {
		token = old.Token
	}
	s, err := domain.NewSettings(in.Enabled, in.ServerURL, in.Topic, token, in.Rules)
	if err != nil {
		return SettingsView{}, err
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save notification settings: %w", err)
	}
	uc.mu.Lock()
	uc.settings = &s
	uc.mu.Unlock()
	return uc.view(s), nil
}

// ErrNotConfigured is returned by SendTest without a topic.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrInvalid, "Please enter and save a topic first.")

// SendTest implements Notifier; it delivers synchronously so errors are shown.
func (uc *notifier) SendTest(ctx context.Context) error {
	s, err := uc.current(ctx)
	if err != nil {
		return err
	}
	if s.Topic == "" {
		return ErrNotConfigured
	}
	err = uc.sender.Send(ctx, s, domain.Message{Title: "trisse-home test", Body: "Notifications work. 🎉", Priority: domain.PriorityDefault, Tags: []string{"tada"}})
	uc.record(err)
	if err != nil {
		var pe *kernel.PublicError
		if errors.As(err, &pe) {
			return err
		}
		return kernel.NewPublicError(kernel.ErrUnavailable, "The ntfy server could not be reached: "+err.Error())
	}
	return nil
}
