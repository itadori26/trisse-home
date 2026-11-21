package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"waylight/internal/modules/proxmox/domain"
	"waylight/internal/shared/kernel"
)

// SettingsView never contains the token secret, only whether one is stored.
type SettingsView struct {
	Configured  bool
	URL         string
	TokenID     string
	HasSecret   bool
	Fingerprint string
}

// SettingsInput is the input of SaveSettings; an empty secret keeps the
// stored one.
type SettingsInput struct {
	URL, TokenID, Secret, Fingerprint string
}

// StatusView is the cached cluster state.
type StatusView struct {
	Cluster   domain.Cluster
	UpdatedAt time.Time
	// Error is set when the last refresh failed; Cluster then holds the
	// last known state.
	Error string
	// Actions lists the guest IDs with start/shutdown enabled.
	Actions []int
}

// connectionTTL is how long the state of a named connection is reused.
// The background job only refreshes the module settings, so connections
// are read on demand at the same pace.
const connectionTTL = 30 * time.Second

// Proxmox is the input port of the module (#84). Status and Act take the
// named connection the widget chose (#610); an empty connectionID, or a
// connection that is not usable, uses the module settings.
type Proxmox interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Status(ctx context.Context, connectionID string) (StatusView, error)
	// Refresh is run by the background job.
	Refresh(ctx context.Context) error
	// Act starts or shuts down an allow-listed guest.
	Act(ctx context.Context, connectionID string, guestID int, action string) error
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

// Client talks to the Proxmox VE API (output port).
type Client interface {
	Cluster(ctx context.Context, s domain.Settings) (domain.Cluster, error)
	Act(ctx context.Context, s domain.Settings, g domain.Guest, a domain.Action) error
}

// Deps bundles the output ports and the action policy.
type Deps struct {
	Repo   SettingsRepository
	Client Client
	// Conns resolves named connections; may be nil.
	Conns  Connections
	Clock  kernel.Clock
	Policy domain.ActionPolicy
	// Audit records power actions in the security event log; may be nil.
	Audit kernel.AuditSink
}

// errUnreachable is shown in the widget; details go to the log.
var errUnreachable = kernel.NewPublicError(kernel.ErrUnavailable, "Proxmox VE is not reachable.")

type proxmox struct {
	d      Deps
	mu     sync.RWMutex
	cached *StatusView
	// byConnection holds the state per named connection.
	byConnection map[string]StatusView
}

// NewProxmox returns the Proxmox use cases.
func NewProxmox(d Deps) Proxmox {
	if d.Audit == nil {
		d.Audit = kernel.DiscardAudit{}
	}
	return &proxmox{d: d, byConnection: map[string]StatusView{}}
}

// Settings implements Proxmox.
func (uc *proxmox) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.d.Repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return settingsView(s), nil
}

// SaveSettings implements Proxmox; the connection is tested before saving.
func (uc *proxmox) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	secret := in.Secret
	if secret == "" {
		if old, err := uc.d.Repo.Load(ctx); err == nil {
			// URL, token ID and pinned certificate identify the target (#191).
			fp, _ := domain.NormaliseFingerprint(in.Fingerprint) // validated by NewSettings below
			if kernel.TargetChanged([]string{old.URL, old.TokenID, old.Fingerprint}, []string{in.URL, in.TokenID, fp}) {
				return SettingsView{}, kernel.SecretRequired("secret")
			}
			secret = old.Secret
		}
	}
	s, err := domain.NewSettings(in.URL, in.TokenID, secret, in.Fingerprint)
	if err != nil {
		return SettingsView{}, err
	}
	cluster, err := uc.d.Client.Cluster(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.SetupFailure("Proxmox VE", err)
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(cluster)
	return settingsView(s), nil
}

// connection returns the settings of a usable named connection.
func (uc *proxmox) connection(ctx context.Context, id string) (domain.Settings, bool) {
	if id == "" || uc.d.Conns == nil {
		return domain.Settings{}, false
	}
	s, err := uc.d.Conns.Settings(ctx, id)
	return s, err == nil
}

// Status implements Proxmox (cache-aside).
func (uc *proxmox) Status(ctx context.Context, connectionID string) (StatusView, error) {
	if s, ok := uc.connection(ctx, connectionID); ok {
		return uc.connectionStatus(ctx, connectionID, s), nil
	}
	if v, ok := uc.cachedView(); ok {
		return v, nil
	}
	if _, err := uc.d.Repo.Load(ctx); err != nil {
		return StatusView{}, err
	}
	_ = uc.Refresh(ctx) // a failed refresh is reported through StatusView.Error
	if v, ok := uc.cachedView(); ok {
		return v, nil
	}
	return StatusView{}, errUnreachable
}

// Refresh implements Proxmox.
func (uc *proxmox) Refresh(ctx context.Context) error {
	s, err := uc.d.Repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	cluster, err := uc.d.Client.Cluster(ctx, s)
	if err != nil {
		uc.mu.Lock()
		v := StatusView{}
		if uc.cached != nil {
			v = *uc.cached
		}
		v.Error = errUnreachable.Error()
		uc.cached = &v
		uc.mu.Unlock()
		return fmt.Errorf("query proxmox: %w", err)
	}
	uc.store(cluster)
	return nil
}

// connectionStatus returns the cached state of a named connection or
// reads it. Like the module state, a failed read keeps the last state and
// reports the error alongside.
func (uc *proxmox) connectionStatus(ctx context.Context, id string, s domain.Settings) StatusView {
	now := uc.d.Clock.Now()
	uc.mu.RLock()
	v, ok := uc.byConnection[id]
	uc.mu.RUnlock()
	if ok && now.Sub(v.UpdatedAt) < connectionTTL {
		return v
	}
	if c, err := uc.d.Client.Cluster(ctx, s); err != nil {
		v.Error, v.UpdatedAt = errUnreachable.Error(), now
	} else {
		v = uc.view(c)
	}
	uc.mu.Lock()
	uc.byConnection[id] = v
	uc.mu.Unlock()
	return v
}

// Act implements Proxmox. The guest's node and kind come from the last
// known cluster state, never from the request.
func (uc *proxmox) Act(ctx context.Context, connectionID string, guestID int, action string) error {
	a := domain.Action(action)
	if err := uc.d.Policy.Check(guestID, a); err != nil {
		return err
	}
	s, viaConnection := uc.connection(ctx, connectionID)
	if !viaConnection {
		var err error
		if s, err = uc.d.Repo.Load(ctx); err != nil {
			return err
		}
	}
	status, err := uc.Status(ctx, connectionID)
	if err != nil {
		return err
	}
	g, ok := status.Cluster.Guest(guestID)
	if !ok {
		return domain.ErrGuestNotFound
	}
	if err := uc.d.Client.Act(ctx, s, g, a); err != nil {
		return kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	uc.d.Audit.Audit(ctx, kernel.AuditEvent{Type: "proxmox_" + string(a), Detail: fmt.Sprintf("%s %d (%s)", g.Kind, g.ID, g.Name)})
	if viaConnection {
		// The next status reads the new state; the old one stays as the last
		// known state should that read fail.
		uc.mu.Lock()
		v := uc.byConnection[connectionID]
		v.UpdatedAt = time.Time{}
		uc.byConnection[connectionID] = v
		uc.mu.Unlock()
		return nil
	}
	_ = uc.Refresh(ctx) // the new state shows up with the next refresh at the latest
	return nil
}

func (uc *proxmox) cachedView() (StatusView, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	if uc.cached == nil {
		return StatusView{}, false
	}
	return *uc.cached, true
}

func (uc *proxmox) store(c domain.Cluster) {
	v := uc.view(c)
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = &v
}

// view is the sorted cluster with the guests whose actions are enabled.
func (uc *proxmox) view(c domain.Cluster) StatusView {
	c.Sort()
	actions := []int{}
	for _, g := range c.Guests {
		if uc.d.Policy.Allows(g.ID) {
			actions = append(actions, g.ID)
		}
	}
	return StatusView{Cluster: c, UpdatedAt: uc.d.Clock.Now(), Actions: actions}
}

func settingsView(s domain.Settings) SettingsView {
	return SettingsView{Configured: true, URL: s.URL, TokenID: s.TokenID, HasSecret: s.Secret != "", Fingerprint: s.Fingerprint}
}
