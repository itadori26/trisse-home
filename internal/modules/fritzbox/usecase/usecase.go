package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/fritzbox/domain"
	"trisse-home/internal/shared/kernel"
)

// callsTTL is how long a fetched call list is reused (cache-aside, ADR 0006).
const callsTTL = 2 * time.Minute

// SettingsView never contains the password.
type SettingsView struct {
	Configured  bool
	URL         string
	Username    string
	HasPassword bool
}

// SettingsInput is the input of SaveSettings; an empty password keeps the
// stored one unless the username is cleared.
type SettingsInput struct {
	URL, Username, Password string
}

// StatusView is the cached status.
type StatusView struct {
	Status    domain.Status
	UpdatedAt time.Time
	Error     string
}

// CallsView is the call list of a calls widget (#168).
type CallsView struct {
	Calls     []domain.Call
	UpdatedAt time.Time
	// Hint explains an empty list (no login configured).
	Hint string
	// Error is set when the box did not answer; Calls are the last known ones.
	Error string
}

// FritzBox is the input port of the module.
type FritzBox interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	Status(ctx context.Context) (StatusView, error)
	// Calls returns the call list, fetched only when asked for (#168).
	Calls(ctx context.Context, f domain.CallFilter) (CallsView, error)
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Box queries the FRITZ!Box (output port).
type Box interface {
	// WAN returns the connection state; the model name is returned as well.
	WAN(ctx context.Context, s domain.Settings) (domain.WAN, string, error)
	// Calls returns the newest calls (needs a login).
	Calls(ctx context.Context, s domain.Settings, limit int) ([]domain.Call, error)
}

type fritzbox struct {
	repo  SettingsRepository
	box   Box
	clock kernel.Clock

	mu      sync.RWMutex
	cached  *StatusView
	calls   []domain.Call
	callsAt time.Time
}

// NewFritzBox returns the FRITZ!Box use cases.
func NewFritzBox(repo SettingsRepository, box Box, clock kernel.Clock) FritzBox {
	return &fritzbox{repo: repo, box: box, clock: clock}
}

// Settings implements FritzBox.
func (uc *fritzbox) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{URL: domain.DefaultURL}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: s.HasLogin()}, nil
}

// SaveSettings implements FritzBox; the connection (and login) is tested.
func (uc *fritzbox) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	password := in.Password
	if password == "" && in.Username != "" {
		if old, err := uc.repo.Load(ctx); err == nil && old.Username == in.Username {
			// Compare the normalised address (default port 49000 added).
			if norm, err := domain.NewSettings(in.URL, in.Username, old.Password); err == nil &&
				kernel.TargetChanged([]string{old.URL}, []string{norm.URL}) {
				return SettingsView{}, kernel.SecretRequired("password")
			}
			password = old.Password
		}
	}
	s, err := domain.NewSettings(in.URL, in.Username, password)
	if err != nil {
		return SettingsView{}, err
	}
	if _, _, err := uc.box.WAN(ctx, s); err != nil {
		return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if s.HasLogin() {
		if _, err := uc.box.Calls(ctx, s, 1); err != nil {
			return SettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
		}
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	uc.cached, uc.calls, uc.callsAt = nil, nil, time.Time{}
	uc.mu.Unlock()
	_ = uc.Refresh(ctx)
	return SettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: s.HasLogin()}, nil
}

// Status implements FritzBox (cache-aside).
func (uc *fritzbox) Status(ctx context.Context) (StatusView, error) {
	uc.mu.RLock()
	cached := uc.cached
	uc.mu.RUnlock()
	if cached != nil {
		return *cached, nil
	}
	if _, err := uc.repo.Load(ctx); err != nil {
		return StatusView{}, err
	}
	_ = uc.Refresh(ctx)
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return *uc.cached, nil
}

// Refresh implements FritzBox; it is run by the background job.
func (uc *fritzbox) Refresh(ctx context.Context) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	now := uc.clock.Now()
	wan, model, wanErr := uc.box.WAN(ctx, s)

	uc.mu.Lock()
	defer uc.mu.Unlock()
	if wanErr != nil {
		view := StatusView{UpdatedAt: now}
		if uc.cached != nil {
			view = *uc.cached
		}
		view.Error = "The FRITZ!Box is not reachable."
		uc.cached = &view
		return fmt.Errorf("query fritzbox: %w", wanErr)
	}
	st := domain.Status{Model: model, WAN: wan}
	if !wan.Known {
		st.Hint = "Enable “Transmit status information over UPnP” on the FRITZ!Box (Home Network → Network → Network Settings) to see the throughput."
	}
	uc.cached = &StatusView{Status: st, UpdatedAt: now}
	return nil
}

// Calls implements FritzBox: the list is fetched when a calls widget asks
// and reused for two minutes; the background job never fetches it (#168).
func (uc *fritzbox) Calls(ctx context.Context, f domain.CallFilter) (CallsView, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return CallsView{}, err
	}
	if !s.HasLogin() {
		return CallsView{Calls: []domain.Call{}, Hint: "Add a FRITZ!Box user in the settings to see the call list."}, nil
	}
	now := uc.clock.Now()
	uc.mu.RLock()
	calls, at := uc.calls, uc.callsAt
	uc.mu.RUnlock()
	view := CallsView{}
	if at.IsZero() || now.Sub(at) >= callsTTL {
		fetched, err := uc.box.Calls(ctx, s, domain.MaxCalls)
		if err != nil {
			view.Error = "The call list is not available."
		} else {
			calls, at = fetched, now
			uc.mu.Lock()
			uc.calls, uc.callsAt = fetched, now
			uc.mu.Unlock()
		}
	}
	view.Calls, view.UpdatedAt = f.Apply(calls), at
	return view, nil
}
