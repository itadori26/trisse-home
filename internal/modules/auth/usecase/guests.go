package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// GuestView describes a guest for the owner's settings.
type GuestView struct {
	ID        string
	Name      string
	Tabs      []string
	Disabled  bool
	CreatedAt time.Time
}

// NewGuestRequest is the input of Guests.Create.
type NewGuestRequest struct {
	Name     string
	Password string
	Tabs     []string
}

// UpdateGuestRequest is the input of Guests.Update; an empty password keeps it.
type UpdateGuestRequest struct {
	ID       string
	Tabs     []string
	Disabled bool
	Password string
}

// Guests manages read-only guest accounts (input port, owner only).
type Guests interface {
	List(ctx context.Context) ([]GuestView, error)
	Create(ctx context.Context, req NewGuestRequest) (GuestView, error)
	Update(ctx context.Context, req UpdateGuestRequest) (GuestView, error)
	Delete(ctx context.Context, id string) error
}

type guests struct{ deps Deps }

// NewGuests returns the guest management use cases.
func NewGuests(deps Deps) Guests { return &guests{deps: deps} }

func ownerOnly(ctx context.Context) error {
	if p := kernel.PrincipalFrom(ctx); p.TabScoped() || p.IsToken() {
		return domain.ErrOwnerOnly
	}
	return nil
}

func guestView(a domain.Account) GuestView {
	return GuestView{ID: a.ID(), Name: a.Name(), Tabs: a.Tabs(), Disabled: a.Disabled(), CreatedAt: a.CreatedAt()}
}

// List implements Guests.
func (uc *guests) List(ctx context.Context) ([]GuestView, error) {
	if err := ownerOnly(ctx); err != nil {
		return nil, err
	}
	all, err := uc.deps.Accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	out := []GuestView{}
	for _, a := range all {
		if a.Role() == domain.RoleGuest {
			out = append(out, guestView(a))
		}
	}
	return out, nil
}

func (uc *guests) hash(raw string) (string, error) {
	password, err := domain.NewPassword(raw)
	if err != nil {
		return "", err
	}
	h, err := uc.deps.Hasher.Hash(password.String())
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return h, nil
}

// Create implements Guests.
func (uc *guests) Create(ctx context.Context, req NewGuestRequest) (GuestView, error) {
	existing, err := uc.List(ctx)
	if err != nil {
		return GuestView{}, err
	}
	if len(existing) >= domain.MaxGuests {
		return GuestView{}, domain.ErrTooManyGuests
	}
	hash, err := uc.hash(req.Password)
	if err != nil {
		return GuestView{}, err
	}
	g, err := domain.NewGuest(uc.deps.IDs.NewID(), req.Name, hash, req.Tabs, uc.deps.Clock.Now())
	if err != nil {
		return GuestView{}, err
	}
	if err := uc.deps.Accounts.Create(ctx, g); err != nil {
		return GuestView{}, err
	}
	uc.deps.Events.record(ctx, domain.EventGuestChanged, g.ID(), domain.Client{}, "created "+g.Name())
	return guestView(g), nil
}

func (uc *guests) guest(ctx context.Context, id string) (domain.Account, error) {
	if err := ownerOnly(ctx); err != nil {
		return domain.Account{}, err
	}
	if id == domain.OwnerID {
		return domain.Account{}, domain.ErrOwnerNotAGuest
	}
	a, err := uc.deps.Accounts.Get(ctx, id)
	if err == nil && a.Role() != domain.RoleGuest {
		return domain.Account{}, domain.ErrGuestNotFound // kiosk displays are managed apart
	}
	return a, err
}

// Update implements Guests. Disabling a guest or changing its password ends
// its sessions.
func (uc *guests) Update(ctx context.Context, req UpdateGuestRequest) (GuestView, error) {
	g, err := uc.guest(ctx, req.ID)
	if err != nil {
		return GuestView{}, err
	}
	now := uc.deps.Clock.Now()
	if err := g.SetGuestAccess(req.Tabs, req.Disabled, now); err != nil {
		return GuestView{}, err
	}
	if req.Password != "" {
		hash, err := uc.hash(req.Password)
		if err != nil {
			return GuestView{}, err
		}
		if err := g.ChangePasswordHash(hash, now); err != nil {
			return GuestView{}, err
		}
	}
	if err := uc.deps.Accounts.Update(ctx, g); err != nil {
		return GuestView{}, fmt.Errorf("update guest: %w", err)
	}
	if req.Disabled || req.Password != "" {
		if err := uc.deps.Sessions.DeleteAllFor(ctx, g.ID()); err != nil {
			return GuestView{}, fmt.Errorf("end guest sessions: %w", err)
		}
	}
	uc.deps.Events.record(ctx, domain.EventGuestChanged, g.ID(), domain.Client{}, "updated "+g.Name())
	return guestView(g), nil
}

// Delete implements Guests.
func (uc *guests) Delete(ctx context.Context, id string) error {
	g, err := uc.guest(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.deps.Sessions.DeleteAllFor(ctx, g.ID()); err != nil {
		return fmt.Errorf("end guest sessions: %w", err)
	}
	if uc.deps.Passkeys != nil {
		if err := uc.deps.Passkeys.DeleteByAccount(ctx, g.ID()); err != nil {
			return err
		}
	}
	if err := uc.deps.Accounts.Delete(ctx, g.ID()); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventGuestChanged, g.ID(), domain.Client{}, "deleted "+g.Name())
	return nil
}

// GuestScope tells which modules serve the widgets on a set of dashboard
// tabs (output port, implemented in the composition root).
type GuestScope interface {
	ModulesForTabs(ctx context.Context, tabs []string) (map[string]bool, error)
}

// Authorizer decides whether the principal may call an endpoint (input
// port used by the request guard).
type Authorizer interface {
	Authorize(ctx context.Context, p kernel.Principal, method, path string) error
}

type authorizer struct{ scope GuestScope }

// NewAuthorizer returns the authorization decision; scope may be nil (then
// guests can only use the self-service endpoints).
func NewAuthorizer(scope GuestScope) Authorizer { return &authorizer{scope: scope} }

// Authorize implements Authorizer: the owner may do everything, guests what
// domain.GuestMayAccess allows.
func (uc *authorizer) Authorize(ctx context.Context, p kernel.Principal, method, path string) error {
	if p.IsToken() {
		if !domain.TokenMayAccess(p.Scopes, method, path) {
			return domain.ErrTokenScope
		}
		return nil
	}
	if !p.TabScoped() {
		return nil
	}
	modules := map[string]bool{}
	if uc.scope != nil {
		m, err := uc.scope.ModulesForTabs(ctx, p.Tabs)
		if err != nil {
			return fmt.Errorf("resolve guest scope: %w", err)
		}
		modules = m
	}
	allowed := domain.GuestMayAccess(method, path, modules)
	if p.IsKiosk() {
		allowed = domain.KioskMayAccess(method, path, modules)
	}
	if !allowed {
		return domain.ErrReadOnly
	}
	return nil
}
