package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

var ctx = context.Background()

func TestSetupAccount(t *testing.T) {
	tests := []struct {
		name    string
		env     *env
		token   string
		pw      string
		wantErr error
	}{
		{"success", newEnv(), "setup-secret", goodPassword, nil},
		{"wrong token", newEnv(), "nope", goodPassword, domain.ErrInvalidSetupToken},
		{"weak password", newEnv(), "setup-secret", "short", kernel.ErrInvalid},
		{"already set up", newEnv().withAccount(), "setup-secret", goodPassword, domain.ErrAlreadySetUp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := &countingObserver{}
			uc := NewSetupAccount(tt.env.deps, staticSetupToken("setup-secret"), obs)
			res, err := uc.Execute(ctx, SetupRequest{SetupToken: tt.token, Password: tt.pw, ClientKey: "ip"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && (res.SessionToken == "" || len(tt.env.sessions.byHash) != 1) {
				t.Fatal("expected a session to be issued")
			}
			if wantCalls := map[bool]int{true: 1, false: 0}[tt.wantErr == nil]; obs.calls != wantCalls {
				t.Fatalf("observer called %d times, want %d", obs.calls, wantCalls)
			}
		})
	}
}

func TestLoginAndRateLimit(t *testing.T) {
	e := newEnv().withAccount()
	uc := NewLogin(e.deps)
	for range 3 {
		if _, err := uc.Execute(ctx, LoginRequest{Password: "wrong", ClientKey: "ip"}); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("expected invalid credentials, got %v", err)
		}
	}
	if _, err := uc.Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"}); !errors.Is(err, domain.ErrTooManyAttempts) {
		t.Fatalf("expected rate limit, got %v", err)
	}
	res, err := uc.Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "other-ip"})
	if err != nil || res.SessionToken == "" || res.CSRFToken == "" {
		t.Fatalf("login from other ip failed: %v", err)
	}
	if res.SessionToken == res.CSRFToken {
		t.Fatal("session and csrf tokens must differ")
	}
}

func TestLoginRequiresSetup(t *testing.T) {
	if _, err := NewLogin(newEnv().deps).Execute(ctx, LoginRequest{Password: goodPassword}); !errors.Is(err, domain.ErrSetupRequired) {
		t.Fatalf("expected ErrSetupRequired, got %v", err)
	}
}

func TestAuthenticateLifecycle(t *testing.T) {
	e := newEnv().withAccount()
	res, err := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	if err != nil {
		t.Fatal(err)
	}
	auth := NewAuthenticate(e.sessions, e.clock, e.deps.TTL)
	got, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: res.SessionToken})
	if err != nil || got.CSRFToken != res.CSRFToken || got.Refreshed {
		t.Fatalf("authenticate: %+v %v", got, err)
	}
	e.clock.now = e.clock.now.Add(2 * time.Hour)
	if got, _ := auth.Execute(ctx, AuthenticateRequest{SessionToken: res.SessionToken}); !got.Refreshed {
		t.Fatal("expected sliding refresh")
	}
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: "forged"}); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("forged token accepted: %v", err)
	}
	e.clock.now = e.clock.now.Add(48 * time.Hour)
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: res.SessionToken}); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("expired session accepted: %v", err)
	}
	if n, _ := NewPurgeExpiredSessions(e.sessions, e.clock).Execute(ctx); n != 1 {
		t.Fatalf("purged %d sessions, want 1", n)
	}
}

func TestStatus(t *testing.T) {
	e := newEnv()
	status := NewGetStatus(e.accounts, e.sessions, e.clock)
	if s, _ := status.Execute(ctx, StatusRequest{}); !s.SetupRequired {
		t.Fatal("expected setup required")
	}
	e.withAccount()
	res, _ := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	s, err := status.Execute(ctx, StatusRequest{SessionToken: res.SessionToken})
	if err != nil || !s.Authenticated || s.CSRFToken != res.CSRFToken {
		t.Fatalf("status: %+v %v", s, err)
	}
	if s, _ := status.Execute(ctx, StatusRequest{SessionToken: "x"}); s.Authenticated || s.SetupRequired {
		t.Fatalf("unexpected status for invalid token: %+v", s)
	}
}

func TestChangePasswordRevokesSessions(t *testing.T) {
	e := newEnv().withAccount()
	old, _ := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	uc := NewChangePassword(e.deps)
	if _, err := uc.Execute(ctx, ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "another-good-password"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	res, err := uc.Execute(ctx, ChangePasswordRequest{CurrentPassword: goodPassword, NewPassword: "another-good-password"})
	if err != nil {
		t.Fatal(err)
	}
	auth := NewAuthenticate(e.sessions, e.clock, e.deps.TTL)
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: old.SessionToken}); err == nil {
		t.Fatal("old session survived password change")
	}
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: res.SessionToken}); err != nil {
		t.Fatalf("new session invalid: %v", err)
	}
}

func TestLogoutAndReset(t *testing.T) {
	e := newEnv().withAccount()
	res, _ := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	if err := NewLogout(e.sessions).Execute(ctx, LogoutRequest{SessionToken: res.SessionToken}); err != nil || len(e.sessions.byHash) != 0 {
		t.Fatalf("logout failed: %v", err)
	}
	if err := NewResetAccount(e.accounts, e.sessions).Execute(ctx); err != nil || e.accounts.account != nil {
		t.Fatalf("reset failed: %v", err)
	}
}

func TestManageSessions(t *testing.T) {
	e := newEnv().withAccount()
	a, _ := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "10.0.0.1", UserAgent: "Firefox"})
	b, _ := NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "10.0.0.2", UserAgent: "Chrome"})
	uc := NewManageSessions(e.sessions, e.clock, nil)
	list, err := uc.List(ctx, a.SessionToken)
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %v %v", list, err)
	}
	var otherID string
	for _, s := range list {
		if s.Current != (s.UserAgent == "Firefox") {
			t.Fatalf("current flag wrong: %+v", s)
		}
		if !s.Current {
			otherID = s.ID
		}
	}
	if err := uc.Revoke(ctx, a.SessionToken, otherID); err != nil {
		t.Fatal(err)
	}
	if _, err := NewAuthenticate(e.sessions, e.clock, e.deps.TTL).Execute(ctx, AuthenticateRequest{SessionToken: b.SessionToken}); err == nil {
		t.Fatal("revoked session still valid")
	}
	_, _ = NewLogin(e.deps).Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "10.0.0.3"})
	if n, _ := uc.RevokeOthers(ctx, a.SessionToken); n != 1 {
		t.Fatalf("revoke others removed %d", n)
	}
}
