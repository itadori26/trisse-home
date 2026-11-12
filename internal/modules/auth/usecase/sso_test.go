package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeOIDC struct {
	id    domain.ExternalIdentity
	err   error
	nonce string
}

func (f *fakeOIDC) AuthCodeURL(_ context.Context, state, nonce, _ string) (string, error) {
	f.nonce = nonce
	return "https://id.example/auth?state=" + state, nil
}

func (f *fakeOIDC) Exchange(_ context.Context, code, _, nonce string) (domain.ExternalIdentity, error) {
	if code != "good" || nonce != f.nonce {
		return domain.ExternalIdentity{}, errors.New("invalid grant")
	}
	return f.id, f.err
}

func TestOIDCLogin(t *testing.T) {
	e := newEnv().withAccount()
	provider := &fakeOIDC{id: domain.ExternalIdentity{Subject: "s1", Email: "me@example.org", EmailVerified: true}}
	uc := NewSSO(e.deps, SSOConfig{OIDC: provider, Mapping: domain.IdentityMapping{Emails: []string{"me@example.org"}}, LocalLogin: true})
	client := domain.Client{IP: "10.0.0.2"}

	_, state, err := uc.BeginOIDC(ctx)
	if err != nil {
		t.Fatal(err)
	}
	res, _, err := uc.CompleteOIDC(ctx, state, "good", client)
	if err != nil || res.SessionToken == "" {
		t.Fatalf("%+v %v", res, err)
	}
	if _, _, err := uc.CompleteOIDC(ctx, state, "good", client); !errors.Is(err, domain.ErrSSOStateInvalid) {
		t.Fatalf("state reused: %v", err)
	}
	_, state, _ = uc.BeginOIDC(ctx)
	if _, _, err := uc.CompleteOIDC(ctx, state, "bad", client); !errors.Is(err, kernel.ErrUnauthorized) {
		t.Fatalf("bad code: %v", err)
	}

	provider.id = domain.ExternalIdentity{Subject: "s2", Email: "intruder@example.org", EmailVerified: true}
	_, state, _ = uc.BeginOIDC(ctx)
	if _, _, err := uc.CompleteOIDC(ctx, state, "good", client); !errors.Is(err, domain.ErrSSONotAllowed) {
		t.Fatalf("unlisted user: %v", err)
	}

	_, state, _ = uc.BeginOIDC(ctx)
	e.clock.now = e.clock.now.Add(11 * time.Minute)
	if _, _, err := uc.CompleteOIDC(ctx, state, "good", client); !errors.Is(err, domain.ErrSSOStateInvalid) {
		t.Fatalf("expired state: %v", err)
	}
}

func TestSSORequiresLocalSecondFactor(t *testing.T) {
	e := newEnv().withAccount()
	owner, _ := e.accounts.Get(ctx, domain.OwnerID)
	_ = owner.StartTOTPEnrollment("SECRET")
	_ = owner.ConfirmTOTP(1, []string{"abcd-efgh"})
	_ = e.accounts.Update(ctx, owner)
	uc := NewSSO(e.deps, SSOConfig{HeaderUsers: []string{"me"}, Require2FA: true, LocalLogin: true})
	client := domain.Client{IP: "10.0.0.3"}
	_, ticket, err := uc.HeaderLogin(ctx, "me", client)
	if !errors.Is(err, domain.ErrSSOSecondFactorNeed) || ticket == "" {
		t.Fatalf("expected a ticket: %v", err)
	}
	if _, err := uc.CompleteSecondFactor(ctx, ticket, "000000", client); err == nil {
		t.Fatal("wrong code accepted")
	}
	res, err := uc.CompleteSecondFactor(ctx, ticket, "abcd-efgh", client)
	if err != nil || res.SessionToken == "" {
		t.Fatalf("%+v %v", res, err)
	}
	if _, _, err := uc.HeaderLogin(ctx, "someone-else", client); !errors.Is(err, domain.ErrSSONotAllowed) {
		t.Fatalf("unlisted header user: %v", err)
	}
}

func TestPasswordLoginCanBeDisabled(t *testing.T) {
	e := newEnv().withAccount()
	e.deps.PasswordLoginDisabled = true
	if _, err := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "x"}); !errors.Is(err, domain.ErrLocalLoginDisabled) {
		t.Fatalf("password login still possible: %v", err)
	}
}
