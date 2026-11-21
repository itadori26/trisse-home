package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/shared/kernel"
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
// stored one. SignIn is a finished sign-in of the setup assistant (#654)
// that replaces URL and token.
type SettingsInput struct {
	URL, Token, SignIn string
}

// EntityView is an entity with what the user may do with it.
type EntityView struct {
	domain.Entity
	// Action is the tile's action; empty when read-only or not allow-listed.
	Action domain.ActionKind
}

// SmartHome is the input port of the module (#49). The methods that talk
// to the hub take the named connection the widget chose (#610); an empty
// connectionID, or a connection that is not usable, uses the module
// settings.
type SmartHome interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Entities lists every entity for the widget's entity picker.
	Entities(ctx context.Context, connectionID string) ([]EntityView, error)
	// States returns the given entities in the given order; unknown IDs are
	// skipped.
	States(ctx context.Context, connectionID string, ids []string) ([]EntityView, error)
	// Act toggles or presses an allow-listed entity.
	Act(ctx context.Context, connectionID, id string) error
	// Compact is the compact state of one entity and the entity itself (#614).
	Compact(ctx context.Context, connectionID, id string) (kernel.CompactState, EntityView, error)
}

// Connections resolves a named connection of the integrations module to
// settings (output port, #610). Any error means the module settings are
// used instead.
type Connections interface {
	Settings(ctx context.Context, id string) (domain.Settings, error)
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

// TokenRenewer renews the access token of a signed-in connection with its
// refresh token (output port). A refused refresh token wraps
// kernel.ErrTargetAuth: the admin has to sign in again.
type TokenRenewer interface {
	Renew(ctx context.Context, s domain.Settings) (domain.Settings, error)
}

// Deps bundles the output ports and the action policy.
type Deps struct {
	Repo     SettingsRepository
	Provider Provider
	// Conns resolves named connections; may be nil.
	Conns Connections
	// Renewer and SignIns serve signing in with Home Assistant (#654);
	// without them only pasted tokens work.
	Renewer TokenRenewer
	SignIns kernel.SignIns
	Clock   kernel.Clock
	Policy  domain.ActionPolicy
	// Audit records actions in the security event log; may be nil.
	Audit kernel.AuditSink
}

type cachedStates struct {
	states  []domain.Entity
	fetched time.Time
}

type smartHome struct {
	d  Deps
	mu sync.Mutex
	// states holds one list per connection ID; "" is the module settings.
	states map[string]cachedStates
	// renewing serialises renewals, so two requests never spend one
	// refresh token at the same time.
	renewing sync.Mutex
}

// NewSmartHome returns the smart home use cases.
func NewSmartHome(d Deps) SmartHome {
	if d.Audit == nil {
		d.Audit = kernel.DiscardAudit{}
	}
	return &smartHome{d: d, states: map[string]cachedStates{}}
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
	s, err := uc.input(ctx, in)
	if err != nil {
		return SettingsView{}, err
	}
	if s, err = uc.renew(ctx, s); err != nil {
		return SettingsView{}, kernel.SetupFailure("Home Assistant", err)
	}
	states, err := uc.d.Provider.States(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.SetupFailure("Home Assistant", err)
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	clear(uc.states)
	uc.states[""] = cachedStates{states: states, fetched: uc.d.Clock.Now()}
	uc.mu.Unlock()
	return SettingsView{Configured: true, URL: s.URL, HasToken: true}, nil
}

// input turns the form into settings: a sign-in, a pasted token, or the
// stored credential when the token is left empty.
func (uc *smartHome) input(ctx context.Context, in SettingsInput) (domain.Settings, error) {
	if in.SignIn != "" {
		g, err := kernel.RedeemSignIn(ctx, uc.d.SignIns, in.SignIn, kernel.SignInHomeAssistant)
		if err != nil {
			return domain.Settings{}, err
		}
		return domain.SignedIn(g.URL, g)
	}
	if in.Token == "" {
		if old, err := uc.d.Repo.Load(ctx); err == nil {
			// The stored token is never sent to another address (#191).
			if kernel.TargetChanged([]string{old.URL}, []string{in.URL}) {
				return domain.Settings{}, kernel.SecretRequired("token")
			}
			s, err := domain.NewSettings(in.URL, old.Token)
			if err != nil {
				return domain.Settings{}, err
			}
			// Keep a signed-in connection signed in.
			s.Refresh, s.ClientID, s.Expiry = old.Refresh, old.ClientID, old.Expiry
			return s, nil
		}
	}
	return domain.NewSettings(in.URL, in.Token)
}

// load returns the settings of the named connection, or the module
// settings with a valid access token while that connection does not exist
// or is not usable. A connection holds a pasted long-lived token: the copy
// skips signed-in settings, so only the module settings are ever renewed.
func (uc *smartHome) load(ctx context.Context, connectionID string) (domain.Settings, error) {
	if connectionID != "" && uc.d.Conns != nil {
		if s, err := uc.d.Conns.Settings(ctx, connectionID); err == nil {
			return s, nil
		}
	}
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return domain.Settings{}, err
	}
	if !s.NeedsRenewal(uc.d.Clock.Now()) {
		return s, nil
	}
	uc.renewing.Lock()
	defer uc.renewing.Unlock()
	// Another request may have renewed it while this one waited.
	if s, err = uc.d.Repo.Load(ctx); err != nil || !s.NeedsRenewal(uc.d.Clock.Now()) {
		return s, err
	}
	if s, err = uc.renew(ctx, s); err != nil {
		e := kernel.NewPublicError(kernel.ErrUnavailable, "Home Assistant did not renew the sign-in. Set up Home Assistant again.")
		e.Cause = err
		return domain.Settings{}, e
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return domain.Settings{}, fmt.Errorf("save renewed token: %w", err)
	}
	return s, nil
}

// renew renews the access token of s when it is due.
func (uc *smartHome) renew(ctx context.Context, s domain.Settings) (domain.Settings, error) {
	if !s.NeedsRenewal(uc.d.Clock.Now()) {
		return s, nil
	}
	if uc.d.Renewer == nil {
		return domain.Settings{}, fmt.Errorf("%w: no token renewer", kernel.ErrTargetAuth)
	}
	renewed, err := uc.d.Renewer.Renew(ctx, s)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("renew the access token: %w", err)
	}
	return renewed, nil
}

// Entities implements SmartHome.
func (uc *smartHome) Entities(ctx context.Context, connectionID string) ([]EntityView, error) {
	states, err := uc.current(ctx, connectionID)
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
func (uc *smartHome) States(ctx context.Context, connectionID string, ids []string) ([]EntityView, error) {
	if len(ids) > MaxEntitiesPerWidget {
		return nil, kernel.NewValidationErrorf("ids", "At most {max} entities per widget.", "max", MaxEntitiesPerWidget)
	}
	states, err := uc.current(ctx, connectionID)
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
func (uc *smartHome) Act(ctx context.Context, connectionID, id string) error {
	if !domain.ValidEntityID(id) {
		return domain.ErrEntityNotFound
	}
	if !uc.d.Policy.Allows(id) {
		return domain.ErrActionNotAllowed
	}
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return err
	}
	states, err := uc.current(ctx, connectionID)
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
func (uc *smartHome) current(ctx context.Context, connectionID string) ([]domain.Entity, error) {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return nil, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.d.Clock.Now()
	if c, ok := uc.states[connectionID]; ok && now.Sub(c.fetched) < statesTTL {
		return c.states, nil
	}
	states, err := uc.d.Provider.States(ctx, s)
	if err != nil {
		return nil, kernel.NewPublicError(kernel.ErrUnavailable, "Home Assistant is not reachable.")
	}
	uc.states[connectionID] = cachedStates{states: states, fetched: now}
	return states, nil
}

func (uc *smartHome) invalidate() {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	// Two connections may point at the same hub, so every cached list goes.
	clear(uc.states)
}

func (uc *smartHome) view(e domain.Entity) EntityView {
	v := EntityView{Entity: e}
	if uc.d.Policy.Allows(e.ID) {
		v.Action, _ = e.Action()
	}
	return v
}
