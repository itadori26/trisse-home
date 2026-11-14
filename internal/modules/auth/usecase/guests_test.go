package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

func TestGuestLifecycle(t *testing.T) {
	e := newEnv().withAccount()
	guests := NewGuests(e.deps)
	g, err := guests.Create(ctx, NewGuestRequest{Name: "anna", Password: "guest-password-1", Tabs: []string{"home"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := guests.Create(ctx, NewGuestRequest{Name: "anna", Password: "guest-password-1"}); !errors.Is(err, domain.ErrGuestNameTaken) {
		t.Fatalf("duplicate: %v", err)
	}

	login := NewLogin(e.deps)
	if _, err := login.Execute(ctx, LoginRequest{Username: "anna", Password: goodPassword, ClientKey: "a"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("the owner's password must not open a guest account: %v", err)
	}
	if _, err := login.Execute(ctx, LoginRequest{Username: "nobody", Password: goodPassword, ClientKey: "b"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("unknown name: %v", err)
	}
	session, err := login.Execute(ctx, LoginRequest{Username: "Anna ", Password: "guest-password-1", ClientKey: "c"})
	if err != nil {
		t.Fatal(err)
	}
	auth := NewAuthenticate(e.sessions, e.accounts, e.clock, e.deps.TTL)
	res, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: session.SessionToken})
	if err != nil || !res.Principal.IsGuest() || res.Principal.AccountID != g.ID || !res.Principal.CanSeeTab("home") || res.Principal.CanSeeTab("server") {
		t.Fatalf("principal: %+v %v", res.Principal, err)
	}

	// A guest cannot manage guests.
	guestCtx := kernel.WithPrincipal(ctx, res.Principal)
	if _, err := NewGuests(e.deps).List(guestCtx); !errors.Is(err, domain.ErrOwnerOnly) {
		t.Fatalf("guest listed guests: %v", err)
	}

	// Disabling ends the session immediately.
	if _, err := guests.Update(ctx, UpdateGuestRequest{ID: g.ID, Tabs: []string{"home"}, Disabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: session.SessionToken}); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("disabled guest still signed in: %v", err)
	}
	if _, err := login.Execute(ctx, LoginRequest{Username: "anna", Password: "guest-password-1", ClientKey: "d"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("disabled guest logged in: %v", err)
	}
	if err := guests.Delete(ctx, domain.OwnerID); !errors.Is(err, domain.ErrOwnerNotAGuest) {
		t.Fatalf("owner deleted as guest: %v", err)
	}
	if err := guests.Delete(ctx, g.ID); err != nil {
		t.Fatal(err)
	}
	if list, _ := guests.List(ctx); len(list) != 0 {
		t.Fatalf("left: %v", list)
	}
}

type fixedScope map[string]bool

func (f fixedScope) ReachForTabs(_ context.Context, _ []string) (domain.GuestReach, error) {
	return domain.GuestReach{Modules: f}, nil
}

func TestAuthorizer(t *testing.T) {
	a := NewAuthorizer(fixedScope{"notes": true})
	guest := kernel.Principal{AccountID: "g1", Role: "guest", Tabs: []string{"home"}}
	if err := a.Authorize(ctx, guest, "GET", "/api/notes"); err != nil {
		t.Fatal(err)
	}
	if err := a.Authorize(ctx, guest, "POST", "/api/notes"); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("guest write: %v", err)
	}
	if err := a.Authorize(ctx, kernel.Principal{AccountID: "owner", Role: "owner"}, "DELETE", "/api/anything"); err != nil {
		t.Fatal(err)
	}
}
