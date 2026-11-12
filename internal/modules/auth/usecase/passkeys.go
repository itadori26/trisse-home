package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/auth/domain"
)

// passkeyChallengeTTL bounds a registration or login ceremony.
const passkeyChallengeTTL = 5 * time.Minute

// maxPendingChallenges bounds memory for unfinished ceremonies.
const maxPendingChallenges = 1000

// PasskeyCredential is a verified WebAuthn credential.
type PasskeyCredential struct {
	ID        []byte
	PublicKey []byte
	Alg       int
	SignCount uint32
}

// WebAuthnVerifier checks WebAuthn ceremonies (output port, ADR 0014).
type WebAuthnVerifier interface {
	RPID() string
	VerifyRegistration(challenge, clientDataJSON, attestationObject []byte) (PasskeyCredential, error)
	VerifyAssertion(challenge []byte, cred PasskeyCredential, clientDataJSON, authenticatorData, signature []byte) (uint32, error)
}

// PasskeyRepository persists passkeys (output port).
type PasskeyRepository interface {
	Create(ctx context.Context, p domain.Passkey) error
	ListByAccount(ctx context.Context, accountID string) ([]domain.Passkey, error)
	FindByCredentialID(ctx context.Context, id []byte) (domain.Passkey, error)
	Update(ctx context.Context, p domain.Passkey) error
	Delete(ctx context.Context, accountID, id string) error
	DeleteByAccount(ctx context.Context, accountID string) error
}

// Challenges holds short-lived, single-use WebAuthn challenges and login
// tickets in memory.
type Challenges struct {
	mu sync.Mutex
	m  map[string]challenge
}

type challenge struct {
	accountID string
	value     []byte
	expires   time.Time
}

// NewChallenges returns an empty challenge store.
func NewChallenges() *Challenges { return &Challenges{m: map[string]challenge{}} }

func (c *Challenges) put(key, accountID string, value []byte, now time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.m {
		if now.After(v.expires) {
			delete(c.m, k)
		}
	}
	if len(c.m) >= maxPendingChallenges {
		return domain.ErrTooManyAttempts
	}
	c.m[key] = challenge{accountID: accountID, value: value, expires: now.Add(passkeyChallengeTTL)}
	return nil
}

// take returns and removes a challenge (single use).
func (c *Challenges) take(key string, now time.Time) (challenge, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.m[key]
	delete(c.m, key)
	return v, ok && !now.After(v.expires)
}

// SecondFactorRequired is returned by Login when the account has passkeys:
// the client signs Challenge with one of Credentials and calls
// Passkeys.FinishLogin with Ticket, or sends a TOTP/recovery code instead.
type SecondFactorRequired struct {
	Ticket      string
	Challenge   []byte
	RPID        string
	Credentials [][]byte
	TOTP        bool
}

func (*SecondFactorRequired) Error() string { return "second factor required" }

// PasskeyView describes a passkey in the settings.
type PasskeyView struct {
	ID         string
	Name       string
	CreatedAt  time.Time
	LastUsedAt time.Time
}

// RegistrationOptions are the parameters for navigator.credentials.create().
type RegistrationOptions struct {
	Challenge []byte
	RPID      string
	UserID    []byte
	UserName  string
	Exclude   [][]byte
}

// PasskeyAssertion is the browser's navigator.credentials.get() result.
type PasskeyAssertion struct {
	CredentialID      []byte
	ClientDataJSON    []byte
	AuthenticatorData []byte
	Signature         []byte
}

// Passkeys manages passkeys as a second factor (input port).
type Passkeys interface {
	Available() bool
	List(ctx context.Context) ([]PasskeyView, error)
	BeginRegistration(ctx context.Context) (RegistrationOptions, error)
	// FinishRegistration returns recovery codes when it issued new ones.
	FinishRegistration(ctx context.Context, name string, clientDataJSON, attestationObject []byte) (PasskeyView, []string, error)
	Delete(ctx context.Context, id string) error
	FinishLogin(ctx context.Context, ticket string, a PasskeyAssertion, client domain.Client) (SessionResponse, error)
}

type passkeys struct{ deps Deps }

// NewPasskeys returns the passkey use cases.
func NewPasskeys(deps Deps) Passkeys { return &passkeys{deps: deps} }

// Available implements Passkeys.
func (uc *passkeys) Available() bool { return uc.deps.WebAuthn != nil && uc.deps.Passkeys != nil }

func passkeyView(p domain.Passkey) PasskeyView {
	return PasskeyView{ID: p.ID, Name: p.Name, CreatedAt: p.CreatedAt, LastUsedAt: p.LastUsedAt}
}

// List implements Passkeys.
func (uc *passkeys) List(ctx context.Context) ([]PasskeyView, error) {
	if !uc.Available() {
		return []PasskeyView{}, nil
	}
	list, err := uc.deps.Passkeys.ListByAccount(ctx, me(ctx))
	if err != nil {
		return nil, fmt.Errorf("list passkeys: %w", err)
	}
	out := make([]PasskeyView, 0, len(list))
	for _, p := range list {
		out = append(out, passkeyView(p))
	}
	return out, nil
}

func (uc *passkeys) randomChallenge() ([]byte, error) {
	t, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return nil, fmt.Errorf("generate challenge: %w", err)
	}
	return []byte(t), nil
}

// BeginRegistration implements Passkeys.
func (uc *passkeys) BeginRegistration(ctx context.Context) (RegistrationOptions, error) {
	if !uc.Available() {
		return RegistrationOptions{}, domain.ErrPasskeysNeedHTTPS
	}
	a, err := uc.deps.Accounts.Get(ctx, me(ctx))
	if err != nil {
		return RegistrationOptions{}, err
	}
	existing, err := uc.deps.Passkeys.ListByAccount(ctx, a.ID())
	if err != nil {
		return RegistrationOptions{}, fmt.Errorf("list passkeys: %w", err)
	}
	if len(existing) >= domain.MaxPasskeys {
		return RegistrationOptions{}, domain.ErrTooManyPasskeys
	}
	ch, err := uc.randomChallenge()
	if err != nil {
		return RegistrationOptions{}, err
	}
	if err := uc.deps.Challenges.put("register:"+a.ID(), a.ID(), ch, uc.deps.Clock.Now()); err != nil {
		return RegistrationOptions{}, err
	}
	opts := RegistrationOptions{Challenge: ch, RPID: uc.deps.WebAuthn.RPID(), UserID: []byte(a.ID()), UserName: a.Name()}
	for _, p := range existing {
		opts.Exclude = append(opts.Exclude, p.CredentialID)
	}
	return opts, nil
}

// FinishRegistration implements Passkeys.
func (uc *passkeys) FinishRegistration(ctx context.Context, name string, clientDataJSON, attestationObject []byte) (PasskeyView, []string, error) {
	if !uc.Available() {
		return PasskeyView{}, nil, domain.ErrPasskeysNeedHTTPS
	}
	now := uc.deps.Clock.Now()
	ch, ok := uc.deps.Challenges.take("register:"+me(ctx), now)
	if !ok {
		return PasskeyView{}, nil, domain.ErrPasskeyChallenge
	}
	cred, err := uc.deps.WebAuthn.VerifyRegistration(ch.value, clientDataJSON, attestationObject)
	if err != nil {
		return PasskeyView{}, nil, domain.ErrPasskeyUnsupported
	}
	p, err := domain.NewPasskey(uc.deps.IDs.NewID(), ch.accountID, cred.ID, cred.PublicKey, cred.Alg, cred.SignCount, name, now)
	if err != nil {
		return PasskeyView{}, nil, err
	}
	if err := uc.deps.Passkeys.Create(ctx, p); err != nil {
		return PasskeyView{}, nil, err
	}
	// The first second factor of an account without TOTP brings recovery codes.
	var codes []string
	a, err := uc.deps.Accounts.Get(ctx, ch.accountID)
	if err != nil {
		return PasskeyView{}, nil, err
	}
	if !a.TOTPEnabled() && !a.HasRecoveryCodes() {
		for range domain.RecoveryCodeCount {
			c, err := uc.deps.OTP.NewRecoveryCode()
			if err != nil {
				return PasskeyView{}, nil, err
			}
			codes = append(codes, c)
		}
		a.IssueRecoveryCodes(codes)
		if err := uc.deps.Accounts.Update(ctx, a); err != nil {
			return PasskeyView{}, nil, fmt.Errorf("store recovery codes: %w", err)
		}
	}
	uc.deps.Events.record(ctx, domain.EventPasskeyAdded, a.ID(), domain.Client{}, p.Name)
	return passkeyView(p), codes, nil
}

// Delete implements Passkeys.
func (uc *passkeys) Delete(ctx context.Context, id string) error {
	if !uc.Available() {
		return domain.ErrPasskeyNotFound
	}
	if err := uc.deps.Passkeys.Delete(ctx, me(ctx), id); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventPasskeyRemoved, me(ctx), domain.Client{}, id)
	return nil
}

// startLogin is called by Login after a correct password when the account
// has passkeys; it returns the challenge as *SecondFactorRequired.
func (d Deps) startPasskeyLogin(a domain.Account, keys []domain.Passkey) error {
	t, err := d.Tokens.NewToken()
	if err != nil {
		return fmt.Errorf("generate ticket: %w", err)
	}
	ch, err := d.Tokens.NewToken()
	if err != nil {
		return fmt.Errorf("generate challenge: %w", err)
	}
	if err := d.Challenges.put("login:"+t, a.ID(), []byte(ch), d.Clock.Now()); err != nil {
		return err
	}
	req := &SecondFactorRequired{Ticket: t, Challenge: []byte(ch), RPID: d.WebAuthn.RPID(), TOTP: a.TOTPEnabled()}
	for _, k := range keys {
		req.Credentials = append(req.Credentials, k.CredentialID)
	}
	return req
}

// passkeysOf lists the account's passkeys; nil when passkeys are off.
func (d Deps) passkeysOf(ctx context.Context, accountID string) ([]domain.Passkey, error) {
	if d.WebAuthn == nil || d.Passkeys == nil {
		return nil, nil
	}
	return d.Passkeys.ListByAccount(ctx, accountID)
}

// FinishLogin implements Passkeys: the second step after the password.
func (uc *passkeys) FinishLogin(ctx context.Context, ticket string, a PasskeyAssertion, client domain.Client) (SessionResponse, error) {
	if !uc.Available() {
		return SessionResponse{}, domain.ErrPasskeysNeedHTTPS
	}
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(client.IP, now) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	ch, ok := uc.deps.Challenges.take("login:"+ticket, now)
	if !ok {
		return SessionResponse{}, domain.ErrPasskeyChallenge
	}
	fail := func(err error) (SessionResponse, error) {
		uc.deps.Events.loginFailed(ctx, ch.accountID, client, uc.deps.Limiter.RecordFailure(client.IP, now))
		return SessionResponse{}, err
	}
	p, err := uc.deps.Passkeys.FindByCredentialID(ctx, a.CredentialID)
	if err != nil || p.AccountID != ch.accountID {
		return fail(domain.ErrInvalidCredentials)
	}
	count, err := uc.deps.WebAuthn.VerifyAssertion(ch.value, PasskeyCredential{ID: p.CredentialID, PublicKey: p.PublicKey, Alg: p.Alg, SignCount: p.SignCount},
		a.ClientDataJSON, a.AuthenticatorData, a.Signature)
	if err != nil {
		return fail(domain.ErrInvalidCredentials)
	}
	if err := p.Use(count, now); err != nil {
		return fail(err)
	}
	if err := uc.deps.Passkeys.Update(ctx, p); err != nil {
		return SessionResponse{}, fmt.Errorf("update passkey: %w", err)
	}
	uc.deps.Limiter.Reset(client.IP)
	uc.deps.Events.loginSucceeded(ctx, ch.accountID, client, "password+passkey")
	return uc.deps.issuer().issue(ctx, ch.accountID, client)
}
