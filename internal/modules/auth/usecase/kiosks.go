package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// KioskRepository persists the links and options of kiosk displays (#50).
type KioskRepository interface {
	Save(ctx context.Context, k domain.Kiosk) error
	Get(ctx context.Context, accountID string) (domain.Kiosk, error)
	// FindByTokenHash returns domain.ErrKioskNotFound for unknown links.
	FindByTokenHash(ctx context.Context, hash domain.TokenHash) (domain.Kiosk, error)
	Delete(ctx context.Context, accountID string) error
}

// KioskView describes a display for the owner's settings.
type KioskView struct {
	ID         string
	Name       string
	Tabs       []string
	Disabled   bool
	Options    domain.KioskOptions
	CreatedAt  time.Time
	LastUsedAt time.Time
}

// KioskInput creates or changes a display.
type KioskInput struct {
	Name          string
	Tabs          []string
	Disabled      bool
	RotateSeconds int
	NightFrom     string
	NightTo       string
}

// Kiosks manages kiosk displays (input port). The management is owner
// only; Login is public (the link is the credential) and Options is for the
// signed-in display itself.
type Kiosks interface {
	List(ctx context.Context) ([]KioskView, error)
	// Create returns the display and its secret link token, shown once.
	Create(ctx context.Context, in KioskInput) (KioskView, string, error)
	Update(ctx context.Context, id string, in KioskInput) (KioskView, error)
	// RenewLink replaces the link; the old one and its sessions stop working.
	RenewLink(ctx context.Context, id string) (string, error)
	Delete(ctx context.Context, id string) error
	Login(ctx context.Context, token, clientKey, userAgent string) (SessionResponse, error)
	// Options are the display options of the signed-in kiosk.
	Options(ctx context.Context) (domain.KioskOptions, error)
}

type kiosks struct {
	deps Deps
	repo KioskRepository
}

// NewKiosks returns the kiosk use cases.
func NewKiosks(deps Deps, repo KioskRepository) Kiosks { return &kiosks{deps: deps, repo: repo} }

func (uc *kiosks) view(a domain.Account, k domain.Kiosk) KioskView {
	return KioskView{ID: a.ID(), Name: a.Name(), Tabs: a.Tabs(), Disabled: a.Disabled(), Options: k.Options, CreatedAt: a.CreatedAt(), LastUsedAt: k.LastUsedAt}
}

func (uc *kiosks) List(ctx context.Context) ([]KioskView, error) {
	if err := ownerOnly(ctx); err != nil {
		return nil, err
	}
	all, err := uc.deps.Accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	out := []KioskView{}
	for _, a := range all {
		if a.Role() != domain.RoleKiosk {
			continue
		}
		k, err := uc.repo.Get(ctx, a.ID())
		if err != nil && !errors.Is(err, domain.ErrKioskNotFound) {
			return nil, err
		}
		out = append(out, uc.view(a, k))
	}
	return out, nil
}

func (uc *kiosks) newLink() (string, domain.TokenHash, error) {
	raw, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return "", "", fmt.Errorf("new token: %w", err)
	}
	return raw, domain.HashToken(raw), nil
}

func (uc *kiosks) Create(ctx context.Context, in KioskInput) (KioskView, string, error) {
	existing, err := uc.List(ctx)
	if err != nil {
		return KioskView{}, "", err
	}
	if len(existing) >= domain.MaxKiosks {
		return KioskView{}, "", domain.ErrTooManyKiosks
	}
	opts, err := domain.NewKioskOptions(in.RotateSeconds, in.NightFrom, in.NightTo)
	if err != nil {
		return KioskView{}, "", err
	}
	a, err := domain.NewKiosk(uc.deps.IDs.NewID(), in.Name, in.Tabs, uc.deps.Clock.Now())
	if err != nil {
		return KioskView{}, "", err
	}
	raw, hash, err := uc.newLink()
	if err != nil {
		return KioskView{}, "", err
	}
	if err := uc.deps.Accounts.Create(ctx, a); err != nil {
		return KioskView{}, "", err
	}
	k := domain.Kiosk{AccountID: a.ID(), TokenHash: hash, Options: opts}
	if err := uc.repo.Save(ctx, k); err != nil {
		_ = uc.deps.Accounts.Delete(ctx, a.ID())
		return KioskView{}, "", fmt.Errorf("save kiosk: %w", err)
	}
	uc.deps.Events.record(ctx, domain.EventGuestChanged, a.ID(), domain.Client{}, "kiosk created "+a.Name())
	return uc.view(a, k), raw, nil
}

// kiosk loads a display for the owner; guests and the owner are not kiosks.
func (uc *kiosks) kiosk(ctx context.Context, id string) (domain.Account, domain.Kiosk, error) {
	if err := ownerOnly(ctx); err != nil {
		return domain.Account{}, domain.Kiosk{}, err
	}
	a, err := uc.deps.Accounts.Get(ctx, id)
	if errors.Is(err, domain.ErrGuestNotFound) || errors.Is(err, domain.ErrAccountNotFound) || (err == nil && a.Role() != domain.RoleKiosk) {
		return domain.Account{}, domain.Kiosk{}, domain.ErrKioskNotFound
	}
	if err != nil {
		return domain.Account{}, domain.Kiosk{}, err
	}
	k, err := uc.repo.Get(ctx, id)
	return a, k, err
}

func (uc *kiosks) Update(ctx context.Context, id string, in KioskInput) (KioskView, error) {
	a, k, err := uc.kiosk(ctx, id)
	if err != nil {
		return KioskView{}, err
	}
	opts, err := domain.NewKioskOptions(in.RotateSeconds, in.NightFrom, in.NightTo)
	if err != nil {
		return KioskView{}, err
	}
	if len(in.Tabs) == 0 {
		return KioskView{}, kernel.NewValidationError("tabs", "Choose at least one tab for the display.")
	}
	if err := a.SetGuestAccess(in.Tabs, in.Disabled, uc.deps.Clock.Now()); err != nil {
		return KioskView{}, err
	}
	if err := uc.deps.Accounts.Update(ctx, a); err != nil {
		return KioskView{}, fmt.Errorf("update kiosk: %w", err)
	}
	k.Options = opts
	if err := uc.repo.Save(ctx, k); err != nil {
		return KioskView{}, fmt.Errorf("save kiosk: %w", err)
	}
	if in.Disabled {
		if err := uc.deps.Sessions.DeleteAllFor(ctx, a.ID()); err != nil {
			return KioskView{}, fmt.Errorf("end kiosk sessions: %w", err)
		}
	}
	uc.deps.Events.record(ctx, domain.EventGuestChanged, a.ID(), domain.Client{}, "kiosk updated "+a.Name())
	return uc.view(a, k), nil
}

func (uc *kiosks) RenewLink(ctx context.Context, id string) (string, error) {
	a, k, err := uc.kiosk(ctx, id)
	if err != nil {
		return "", err
	}
	raw, hash, err := uc.newLink()
	if err != nil {
		return "", err
	}
	k.TokenHash = hash
	if err := uc.repo.Save(ctx, k); err != nil {
		return "", fmt.Errorf("save kiosk: %w", err)
	}
	if err := uc.deps.Sessions.DeleteAllFor(ctx, a.ID()); err != nil {
		return "", fmt.Errorf("end kiosk sessions: %w", err)
	}
	uc.deps.Events.record(ctx, domain.EventGuestChanged, a.ID(), domain.Client{}, "kiosk link renewed "+a.Name())
	return raw, nil
}

func (uc *kiosks) Delete(ctx context.Context, id string) error {
	a, _, err := uc.kiosk(ctx, id)
	if err != nil && !errors.Is(err, domain.ErrKioskNotFound) {
		return err
	}
	if a.ID() == "" {
		return domain.ErrKioskNotFound
	}
	if err := uc.deps.Sessions.DeleteAllFor(ctx, a.ID()); err != nil {
		return fmt.Errorf("end kiosk sessions: %w", err)
	}
	if err := uc.repo.Delete(ctx, a.ID()); err != nil {
		return err
	}
	if err := uc.deps.Accounts.Delete(ctx, a.ID()); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventGuestChanged, a.ID(), domain.Client{}, "kiosk deleted "+a.Name())
	return nil
}

// Login implements Kiosks: the link token is the credential. Failed
// attempts count against the same limiter as password logins.
func (uc *kiosks) Login(ctx context.Context, token, clientKey, userAgent string) (SessionResponse, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(clientKey, now) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	client := domain.NewClient(userAgent, clientKey)
	k, err := uc.repo.FindByTokenHash(ctx, domain.HashToken(token))
	var a domain.Account
	if err == nil {
		a, err = uc.deps.Accounts.Get(ctx, k.AccountID)
	}
	if err != nil || a.Role() != domain.RoleKiosk || a.Disabled() {
		if err != nil && !errors.Is(err, domain.ErrKioskNotFound) && !errors.Is(err, domain.ErrGuestNotFound) && !errors.Is(err, domain.ErrAccountNotFound) {
			return SessionResponse{}, err
		}
		uc.deps.Limiter.RecordFailure(clientKey, now)
		return SessionResponse{}, domain.ErrInvalidKioskLink
	}
	k.LastUsedAt, k.LastUsedIP = now, client.IP
	if err := uc.repo.Save(ctx, k); err != nil {
		return SessionResponse{}, fmt.Errorf("save kiosk: %w", err)
	}
	uc.deps.Events.loginSucceeded(ctx, a.ID(), client, "kiosk link")
	return uc.deps.issuer().issue(ctx, a.ID(), client)
}

func (uc *kiosks) Options(ctx context.Context) (domain.KioskOptions, error) {
	p := kernel.PrincipalFrom(ctx)
	if !p.IsKiosk() {
		return domain.KioskOptions{}, domain.ErrKioskNotFound
	}
	k, err := uc.repo.Get(ctx, p.AccountID)
	if err != nil {
		return domain.KioskOptions{}, err
	}
	return k.Options, nil
}
