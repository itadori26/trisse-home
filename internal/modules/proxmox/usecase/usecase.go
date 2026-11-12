package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/proxmox/domain"
	"trisse-home/internal/shared/kernel"
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

// Proxmox is the input port of the module (#84).
type Proxmox interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Status(ctx context.Context) (StatusView, error)
	// Refresh is run by the background job.
	Refresh(ctx context.Context) error
	// Act starts or shuts down an allow-listed guest.
	Act(ctx context.Context, guestID int, action string) error
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
}

// NewProxmox returns the Proxmox use cases.
func NewProxmox(d Deps) Proxmox {
	if d.Audit == nil {
		d.Audit = kernel.DiscardAudit{}
	}
	return &proxmox{d: d}
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
			secret = old.Secret
		}
	}
	s, err := domain.NewSettings(in.URL, in.TokenID, secret, in.Fingerprint)
	if err != nil {
		return SettingsView{}, err
	}
	cluster, err := uc.d.Client.Cluster(ctx, s)
	if err != nil {
		return SettingsView{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not connect to Proxmox VE: {detail}", "detail", err.Error())
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.store(cluster)
	return settingsView(s), nil
}

// Status implements Proxmox (cache-aside).
func (uc *proxmox) Status(ctx context.Context) (StatusView, error) {
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

// Act implements Proxmox. The guest's node and kind come from the last
// known cluster state, never from the request.
func (uc *proxmox) Act(ctx context.Context, guestID int, action string) error {
	a := domain.Action(action)
	if err := uc.d.Policy.Check(guestID, a); err != nil {
		return err
	}
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return err
	}
	status, err := uc.Status(ctx)
	if err != nil {
		return err
	}
	g, ok := status.Cluster.Guest(guestID)
	if !ok {
		return domain.ErrGuestNotFound
	}
	if err := uc.d.Client.Act(ctx, s, g, a); err != nil {
		return kernel.NewPublicErrorf(kernel.ErrUnavailable, "Proxmox VE rejected the action: {detail}", "detail", err.Error())
	}
	uc.d.Audit.Audit(ctx, kernel.AuditEvent{Type: "proxmox_" + string(a), Detail: fmt.Sprintf("%s %d (%s)", g.Kind, g.ID, g.Name)})
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
	c.Sort()
	actions := []int{}
	for _, g := range c.Guests {
		if uc.d.Policy.Allows(g.ID) {
			actions = append(actions, g.ID)
		}
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.cached = &StatusView{Cluster: c, UpdatedAt: uc.d.Clock.Now(), Actions: actions}
}

func settingsView(s domain.Settings) SettingsView {
	return SettingsView{Configured: true, URL: s.URL, TokenID: s.TokenID, HasSecret: s.Secret != "", Fingerprint: s.Fingerprint}
}
