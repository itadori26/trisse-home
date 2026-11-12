package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/auth/domain"
)

type memPasskeys struct{ list []domain.Passkey }

func (m *memPasskeys) Create(_ context.Context, p domain.Passkey) error {
	m.list = append(m.list, p)
	return nil
}
func (m *memPasskeys) ListByAccount(_ context.Context, acc string) ([]domain.Passkey, error) {
	var out []domain.Passkey
	for _, p := range m.list {
		if p.AccountID == acc {
			out = append(out, p)
		}
	}
	return out, nil
}
func (m *memPasskeys) FindByCredentialID(_ context.Context, id []byte) (domain.Passkey, error) {
	for _, p := range m.list {
		if bytes.Equal(p.CredentialID, id) {
			return p, nil
		}
	}
	return domain.Passkey{}, domain.ErrPasskeyNotFound
}
func (m *memPasskeys) Update(_ context.Context, p domain.Passkey) error {
	for i := range m.list {
		if m.list[i].ID == p.ID {
			m.list[i] = p
		}
	}
	return nil
}
func (m *memPasskeys) Delete(_ context.Context, acc, id string) error {
	for i, p := range m.list {
		if p.ID == id && p.AccountID == acc {
			m.list = append(m.list[:i], m.list[i+1:]...)
			return nil
		}
	}
	return domain.ErrPasskeyNotFound
}
func (m *memPasskeys) DeleteByAccount(context.Context, string) error { return nil }

// fakeWebAuthn accepts a signature that equals the challenge and returns
// the counter encoded in the authenticator data's first byte.
type fakeWebAuthn struct{}

func (fakeWebAuthn) RPID() string { return "home.example.org" }
func (fakeWebAuthn) VerifyRegistration(challenge, cd, _ []byte) (PasskeyCredential, error) {
	if !bytes.Equal(challenge, cd) {
		return PasskeyCredential{}, errors.New("challenge")
	}
	return PasskeyCredential{ID: []byte("cred-1"), PublicKey: []byte("pk"), Alg: -7}, nil
}
func (fakeWebAuthn) VerifyAssertion(challenge []byte, _ PasskeyCredential, _, ad, sig []byte) (uint32, error) {
	if !bytes.Equal(challenge, sig) {
		return 0, errors.New("signature")
	}
	return uint32(ad[0]), nil
}

func TestPasskeyAsSecondFactor(t *testing.T) {
	e := newEnv().withAccount()
	e.deps.Passkeys, e.deps.WebAuthn, e.deps.Challenges = &memPasskeys{}, fakeWebAuthn{}, NewChallenges()
	uc := NewPasskeys(e.deps)
	opts, err := uc.BeginRegistration(ctx)
	if err != nil || opts.RPID != "home.example.org" {
		t.Fatalf("%+v %v", opts, err)
	}
	view, codes, err := uc.FinishRegistration(ctx, "YubiKey", opts.Challenge, nil)
	if err != nil || view.Name != "YubiKey" || len(codes) != domain.RecoveryCodeCount {
		t.Fatalf("%+v %d %v", view, len(codes), err)
	}
	if _, _, err := uc.FinishRegistration(ctx, "again", opts.Challenge, nil); !errors.Is(err, domain.ErrPasskeyChallenge) {
		t.Fatalf("registration challenge reused: %v", err)
	}

	login := NewLogin(e.deps)
	_, err = login.Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	var sf *SecondFactorRequired
	if !errors.As(err, &sf) || sf.Ticket == "" || len(sf.Credentials) != 1 || sf.TOTP {
		t.Fatalf("second factor not requested: %v", err)
	}
	assertion := PasskeyAssertion{CredentialID: []byte("cred-1"), AuthenticatorData: []byte{1}, Signature: sf.Challenge}
	res, err := uc.FinishLogin(ctx, sf.Ticket, assertion, domain.Client{IP: "ip"})
	if err != nil || res.SessionToken == "" {
		t.Fatalf("%+v %v", res, err)
	}
	if _, err := uc.FinishLogin(ctx, sf.Ticket, assertion, domain.Client{IP: "ip"}); !errors.Is(err, domain.ErrPasskeyChallenge) {
		t.Fatalf("ticket reused: %v", err)
	}

	// A cloned authenticator replays an old counter.
	_, err = login.Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip"})
	errors.As(err, &sf)
	old := PasskeyAssertion{CredentialID: []byte("cred-1"), AuthenticatorData: []byte{1}, Signature: sf.Challenge}
	if _, err := uc.FinishLogin(ctx, sf.Ticket, old, domain.Client{IP: "ip2"}); !errors.Is(err, domain.ErrPasskeyCloned) {
		t.Fatalf("counter regression accepted: %v", err)
	}

	// Without TOTP a six-digit code never works, a recovery code does.
	if _, err := login.Execute(ctx, LoginRequest{Password: goodPassword, TOTPCode: "000001", ClientKey: "ip3"}); !errors.Is(err, domain.ErrInvalidTOTPCode) {
		t.Fatalf("totp code without totp: %v", err)
	}
	if res, err := login.Execute(ctx, LoginRequest{Password: goodPassword, TOTPCode: codes[0], ClientKey: "ip4"}); err != nil || res.SessionToken == "" {
		t.Fatalf("recovery code: %v", err)
	}
	if err := uc.Delete(ctx, view.ID); err != nil {
		t.Fatal(err)
	}
	if res, err := login.Execute(ctx, LoginRequest{Password: goodPassword, ClientKey: "ip5"}); err != nil || res.SessionToken == "" {
		t.Fatalf("login without second factor after removing the passkey: %v", err)
	}
}
