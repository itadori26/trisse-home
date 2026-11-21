package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/shared/kernel"
)

type oneSignIn struct {
	id    string
	grant kernel.SignInGrant
}

func (o *oneSignIn) Redeem(_ context.Context, id, service string) (kernel.SignInGrant, error) {
	if o.id == "" || id != o.id || service != kernel.SignInHomeAssistant {
		return kernel.SignInGrant{}, kernel.ErrSignInUnknown
	}
	o.id = ""
	return o.grant, nil
}

// fakeRenewer hands out the access token "token" and counts renewals.
type fakeRenewer struct {
	err   error
	calls int
	clock *movingClock
}

func (f *fakeRenewer) Renew(_ context.Context, s domain.Settings) (domain.Settings, error) {
	f.calls++
	if f.err != nil {
		return domain.Settings{}, f.err
	}
	s.Token, s.Expiry = "token", f.clock.now.Add(30*time.Minute)
	return s, nil
}

func signedIn(t *testing.T) (SmartHome, *memRepo, *fakeRenewer, *movingClock, *oneSignIn) {
	t.Helper()
	repo := &memRepo{}
	c := &movingClock{now: time.Unix(1_700_000_000, 0)}
	r := &fakeRenewer{clock: c}
	signIns := &oneSignIn{id: "s1", grant: kernel.SignInGrant{URL: "http://ha:8123", Access: "token", Secret: "refresh",
		ClientID: "https://home.example/", Expiry: c.now.Add(30 * time.Minute)}}
	p := &fakeProvider{states: []domain.Entity{{ID: "sensor.power", Name: "Power", State: "42"}}}
	uc := NewSmartHome(Deps{Repo: repo, Provider: p, Clock: c, Renewer: r, SignIns: signIns})
	return uc, repo, r, c, signIns
}

// TestSaveWithSignIn: the grant of a sign-in is stored with its refresh
// token and client ID, once.
func TestSaveWithSignIn(t *testing.T) {
	ctx := context.Background()
	uc, repo, r, _, _ := signedIn(t)
	v, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://elsewhere", SignIn: "s1"})
	if err != nil || v.URL != "http://ha:8123" {
		t.Fatalf("save: %+v %v", v, err)
	}
	if repo.s.Refresh != "refresh" || repo.s.ClientID != "https://home.example/" || repo.s.Token != "token" || r.calls != 0 {
		t.Fatalf("stored %+v, renewals %d", repo.s, r.calls)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{SignIn: "s1"}); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("a sign-in is redeemed once: %v", err)
	}
	// Saving the form without a token keeps the connection signed in.
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ha:8123/"}); err != nil || repo.s.Refresh != "refresh" {
		t.Fatalf("keep: %+v %v", repo.s, err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://other:8123"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("the stored token must not go to another address: %v", err)
	}
	// A pasted long-lived token ends the sign-in.
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://ha:8123", Token: "token"}); err != nil || repo.s.Refresh != "" {
		t.Fatalf("pasted token: %+v %v", repo.s, err)
	}
}

// TestRenewal: an access token is renewed shortly before it expires, the
// new one is stored, and a refused renewal is reported, not retried
// silently with the old token.
func TestRenewal(t *testing.T) {
	ctx := context.Background()
	uc, repo, r, c, _ := signedIn(t)
	if _, err := uc.SaveSettings(ctx, SettingsInput{SignIn: "s1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Entities(ctx, ""); err != nil || r.calls != 0 {
		t.Fatalf("a fresh token is used as it is: %v, %d renewals", err, r.calls)
	}
	c.now = c.now.Add(29*time.Minute + statesTTL)
	repo.s.Token = "expired"
	if _, err := uc.Entities(ctx, ""); err != nil || r.calls != 1 {
		t.Fatalf("renew before expiry: %v, %d renewals", err, r.calls)
	}
	if repo.s.Token != "token" || !repo.s.Expiry.After(c.now) {
		t.Fatalf("renewed token stored: %+v", repo.s)
	}
	c.now = c.now.Add(time.Hour)
	r.err = kernel.ErrTargetAuth
	if _, err := uc.Entities(ctx, ""); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("refused renewal: %v", err)
	}
}

func TestNeedsRenewal(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cases := map[string]struct {
		s    domain.Settings
		want bool
	}{
		"pasted token":     {domain.Settings{Token: "t"}, false},
		"fresh":            {domain.Settings{Token: "t", Refresh: "r", Expiry: now.Add(10 * time.Minute)}, false},
		"about to expire":  {domain.Settings{Token: "t", Refresh: "r", Expiry: now.Add(30 * time.Second)}, true},
		"expired":          {domain.Settings{Token: "t", Refresh: "r", Expiry: now.Add(-time.Second)}, true},
		"unknown lifetime": {domain.Settings{Token: "t", Refresh: "r"}, true},
	}
	for name, tc := range cases {
		if got := tc.s.NeedsRenewal(now); got != tc.want {
			t.Errorf("%s: got %v", name, got)
		}
	}
}
