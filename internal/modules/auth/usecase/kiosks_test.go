package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

type memKiosks struct{ m map[string]domain.Kiosk }

func (r *memKiosks) Save(_ context.Context, k domain.Kiosk) error { r.m[k.AccountID] = k; return nil }
func (r *memKiosks) Get(_ context.Context, id string) (domain.Kiosk, error) {
	k, ok := r.m[id]
	if !ok {
		return domain.Kiosk{}, domain.ErrKioskNotFound
	}
	return k, nil
}
func (r *memKiosks) FindByTokenHash(_ context.Context, h domain.TokenHash) (domain.Kiosk, error) {
	for _, k := range r.m {
		if k.TokenHash == h {
			return k, nil
		}
	}
	return domain.Kiosk{}, domain.ErrKioskNotFound
}
func (r *memKiosks) Delete(_ context.Context, id string) error { delete(r.m, id); return nil }

func TestKioskLifecycle(t *testing.T) {
	e := newEnv().withAccount()
	repo := &memKiosks{m: map[string]domain.Kiosk{}}
	kiosks := NewKiosks(e.deps, repo)
	if _, _, err := kiosks.Create(ctx, KioskInput{Name: "flur", Tabs: []string{"home"}, RotateSeconds: 3}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("rotation too fast", err)
	}
	k, token, err := kiosks.Create(ctx, KioskInput{Name: "flur", Tabs: []string{"home"}, RotateSeconds: 60, NightFrom: "22:00", NightTo: "06:00"})
	if err != nil || token == "" || k.Options.RotateSeconds != 60 {
		t.Fatalf("%+v %q %v", k, token, err)
	}
	if repo.m[k.ID].TokenHash == domain.TokenHash(token) {
		t.Fatal("the link must be stored hashed")
	}
	// Kiosks are not guests and have no password.
	if gs, _ := NewGuests(e.deps).List(ctx); len(gs) != 0 {
		t.Fatalf("kiosk listed as guest: %+v", gs)
	}
	if _, err := NewGuests(e.deps).Update(ctx, UpdateGuestRequest{ID: k.ID, Tabs: []string{"home", "server"}}); !errors.Is(err, domain.ErrGuestNotFound) {
		t.Fatal("guest endpoints must not change a kiosk", err)
	}
	if _, err := NewLogin(e.deps).Execute(ctx, LoginRequest{Username: "flur", Password: "!kiosk", ClientKey: "x"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatal("a kiosk cannot sign in with a name", err)
	}

	if _, err := kiosks.Login(ctx, "wrong", "c1", "UA"); !errors.Is(err, domain.ErrInvalidKioskLink) {
		t.Fatal(err)
	}
	s, err := kiosks.Login(ctx, token, "c2", "Tablet")
	if err != nil {
		t.Fatal(err)
	}
	auth := NewAuthenticate(e.sessions, e.accounts, e.clock, e.deps.TTL)
	res, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: s.SessionToken})
	if err != nil || !res.Principal.IsKiosk() || !res.Principal.TabScoped() || !res.Principal.CanSeeTab("home") || res.Principal.CanSeeTab("server") {
		t.Fatalf("%+v %v", res.Principal, err)
	}
	kctx := kernel.WithPrincipal(ctx, res.Principal)
	if o, err := kiosks.Options(kctx); err != nil || o.NightFrom != "22:00" {
		t.Fatal(o, err)
	}
	if _, err := kiosks.List(kctx); !errors.Is(err, domain.ErrOwnerOnly) {
		t.Fatal("a kiosk cannot manage kiosks", err)
	}
	if _, err := kiosks.Options(ctx); err == nil {
		t.Fatal("options are for kiosk sessions")
	}
	az := NewAuthorizer(nil)
	if err := az.Authorize(ctx, res.Principal, "PUT", "/api/auth/password"); err == nil {
		t.Fatal("no self-service for kiosks")
	}
	if err := az.Authorize(ctx, res.Principal, "GET", "/api/auth/kiosk"); err != nil {
		t.Fatal(err)
	}

	// A renewed link ends the old one and its sessions.
	fresh, err := kiosks.RenewLink(ctx, k.ID)
	if err != nil || fresh == token {
		t.Fatal(err)
	}
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: s.SessionToken}); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatal("session survived the renewal", err)
	}
	if _, err := kiosks.Login(ctx, token, "c3", "UA"); !errors.Is(err, domain.ErrInvalidKioskLink) {
		t.Fatal("old link still works")
	}
	// Disabled displays cannot sign in.
	if _, err := kiosks.Update(ctx, k.ID, KioskInput{Tabs: []string{"home"}, Disabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := kiosks.Login(ctx, fresh, "c4", "UA"); !errors.Is(err, domain.ErrInvalidKioskLink) {
		t.Fatal("disabled kiosk signed in", err)
	}
	if err := kiosks.Delete(ctx, k.ID); err != nil || len(repo.m) != 0 {
		t.Fatal(err)
	}
	if err := kiosks.Delete(ctx, domain.OwnerID); !errors.Is(err, domain.ErrKioskNotFound) {
		t.Fatal("the owner is not a kiosk", err)
	}
}
