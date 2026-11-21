package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/shared/kernel"
)

// oneSignIn hands out one grant for one ID and service, once.
type oneSignIn struct {
	id, service string
	grant       kernel.SignInGrant
}

func (o *oneSignIn) Redeem(_ context.Context, id, service string) (kernel.SignInGrant, error) {
	if o.id == "" || id != o.id || service != o.service {
		return kernel.SignInGrant{}, kernel.ErrSignInUnknown
	}
	o.id = ""
	return o.grant, nil
}

// TestSaveWithSignIn: a finished sign-in replaces address, user and app
// password, works once, and nothing is saved without it.
func TestSaveWithSignIn(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{}
	signIns := &oneSignIn{id: "s1", service: kernel.SignInNextcloud,
		grant: kernel.SignInGrant{URL: "https://cloud.example", Username: "anna", Secret: "app-pw"}}
	uc := NewNextcloud(repo, src, nil, clk{}, signIns)

	v, err := uc.SaveSettings(ctx, SettingsInput{URL: "https://other.example", SignIn: "s1"})
	if err != nil || v.URL != "https://cloud.example" || repo.s.Username != "anna" || repo.s.AppPassword != "app-pw" {
		t.Fatalf("save with sign-in: %+v %v", repo.s, err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{SignIn: "s1"}); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("a sign-in is redeemed once: %v", err)
	}
	if _, err := NewNextcloud(&memRepo{}, src, nil, clk{}, nil).SaveSettings(ctx, SettingsInput{SignIn: "s1"}); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("without the sign-in module: %v", err)
	}
}
