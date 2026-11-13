package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"trisse-home/internal/modules/smarthome/domain"
	"trisse-home/internal/shared/kernel"
)

// statesTTL is how long fetched states are reused. States are fetched on
// demand only, so nothing is polled while no widget is open.
const statesTTL = 10 * time.Second

// MaxEntitiesPerWidget bounds a widget's entity list.
const MaxEntitiesPerWidget = 50

// SettingsView never contains the token, only whether one is stored.
type SettingsView struct {
	Configured bool
	URL        string
	HasToken   bool
}

// SettingsInput is the input of SaveSettings; an empty token keeps the
// stored one.
type SettingsInput struct {
	URL, Token string
}

// EntityView is an entity with what the user may do with it.
type EntityView struct {
	domain.Entity
	// Action is the tile's action; empty when read-only or not allow-listed.
	Action domain.ActionKind
}

// SmartHome is the input port of the module (#49).
type SmartHome interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Entities lists every entity for the widget's entity picker.
	Entities(ctx context.Context) ([]EntityView, error)
	// States returns the given entities in the given order; unknown IDs are
	// skipped.
	States(ctx context.Context, ids []string) ([]EntityView, error)
	// Act toggles or presses an allow-listed entity.
	Act(ctx context.Context, id string) error
}

// SettingsRepository persists the settings (output port). Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Provider reads states and calls services of a smart home hub (output
// port); Home Assistant's REST API is the first adapter.
type Provider interface {
	States(ctx context.Context, s domain.Settings) ([]domain.Entity, error)
	Call(ctx context.Context, s domain.Settings, svc domain.Service, entityID string) error
}

// Deps bundles the output ports and the action policy.
type Deps struct {
	Repo     SettingsRepository
	Provider Provider
	Clock    kernel.Clock
	Policy   domain.ActionPolicy
	// Audit records actions in the security event log; may be nil.
	Audit kernel.AuditSink
}

type smartHome struct {
	d       Deps
	mu      sync.Mutex
	states  []domain.Entity
	fetched time.Time
}

// NewSmartHome returns the smart home use cases.
func NewSmartHome(d Deps) SmartHome {
	if d.Audit == nil {
		d.Audit = kernel.DiscardAudit{}
	}
	return &smartHome{d: d}
}

// Settings implements SmartHome.
func (uc *smartHome) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.d.Repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, URL: s.URL, HasToken: s.Token != ""}, nil
}

// SaveSettings implements SmartHome; the connection is tested first.
func (uc *smartHome) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	token := in.Token
	if token == "" {
		if old, err := uc.d.Repo.Load(ctx); err == nil {
			token = old.Token
		}
	}
	s, err := domain.NewSettings(in.URL, token)
	if err != nil {
		return SettingsView{}, err
	}
	states, err := uc.d.Provider.States(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not connect to Home Assistant: {detail}", "detail", err.Error())
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	uc.states, uc.fetched = states, uc.d.Clock.Now()
	uc.mu.Unlock()
	return SettingsView{Configured: true, URL: s.URL, HasToken: true}, nil
}

// Entities implements SmartHome.
func (uc *smartHome) Entities(ctx context.Context) ([]EntityView, error) {
	states, err := uc.current(ctx)
	if err != nil {
		return nil, err
	}
	sorted := slices.Clone(states)
	domain.SortByName(sorted)
	out := make([]EntityView, 0, len(sorted))
	for _, e := range sorted {
		out = append(out, uc.view(e))
	}
	return out, nil
}

// States implements SmartHome.
func (uc *smartHome) States(ctx context.Context, ids []string) ([]EntityView, error) {
	if len(ids) > MaxEntitiesPerWidget {
		return nil, kernel.NewValidationErrorf("ids", "At most {max} entities per widget.", "max", MaxEntitiesPerWidget)
	}
	states, err := uc.current(ctx)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]domain.Entity, len(states))
	for _, e := range states {
		byID[e.ID] = e
	}
	out := make([]EntityView, 0, len(ids))
	for _, id := range ids {
		if e, ok := byID[id]; ok {
			out = append(out, uc.view(e))
		}
	}
	return out, nil
}

// Act implements SmartHome.
func (uc *smartHome) Act(ctx context.Context, id string) error {
	if !domain.ValidEntityID(id) {
		return domain.ErrEntityNotFound
	}
	if !uc.d.Policy.Allows(id) {
		return domain.ErrActionNotAllowed
	}
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return err
	}
	states, err := uc.current(ctx)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(states, func(e domain.Entity) bool { return e.ID == id })
	if i < 0 {
		return domain.ErrEntityNotFound
	}
	svc, err := uc.d.Policy.Check(states[i])
	if err != nil {
		return err
	}
	if err := uc.d.Provider.Call(ctx, s, svc, id); err != nil {
		return kernel.NewPublicErrorf(kernel.ErrUnavailable, "Home Assistant rejected the action: {detail}", "detail", err.Error())
	}
	uc.d.Audit.Audit(ctx, kernel.AuditEvent{Type: "smarthome_action", Detail: svc.Domain + "." + svc.Name + " " + id})
	uc.invalidate() // the next read shows the new state
	return nil
}

// current returns cached states younger than statesTTL or fetches them.
func (uc *smartHome) current(ctx context.Context) ([]domain.Entity, error) {
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return nil, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.d.Clock.Now()
	if uc.states != nil && now.Sub(uc.fetched) < statesTTL {
		return uc.states, nil
	}
	states, err := uc.d.Provider.States(ctx, s)
	if err != nil {
		return nil, kernel.NewPublicError(kernel.ErrUnavailable, "Home Assistant is not reachable.")
	}
	uc.states, uc.fetched = states, now
	return states, nil
}

func (uc *smartHome) invalidate() {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.states = nil
}

func (uc *smartHome) view(e domain.Entity) EntityView {
	v := EntityView{Entity: e}
	if uc.d.Policy.Allows(e.ID) {
		v.Action, _ = e.Action()
	}
	return v
}
