package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// ssoStateTTL bounds the time between leaving for the provider and coming back.
const ssoStateTTL = 10 * time.Minute

// maxPendingSSO bounds memory for started but unfinished sign-ins.
const maxPendingSSO = 1000

// OIDCProvider is the OpenID Connect relying party (output port).
type OIDCProvider interface {
	AuthCodeURL(ctx context.Context, state, nonce, verifier string) (string, error)
	// Exchange redeems the code and returns the verified identity.
	Exchange(ctx context.Context, code, verifier, nonce string) (domain.ExternalIdentity, error)
}

// SSOConfig configures single sign-on for the owner.
type SSOConfig struct {
	// OIDC is nil when OpenID Connect is not configured.
	OIDC    OIDCProvider
	Mapping domain.IdentityMapping
	// HeaderUsers may sign in via the trusted proxy header; empty disables it.
	HeaderUsers []string
	// LocalLogin allows the owner's password login (guests always can).
	LocalLogin bool
	// Require2FA asks for the local TOTP code after SSO when it is enabled.
	Require2FA bool
}

// SSOStatus tells the login page which methods exist.
type SSOStatus struct {
	OIDC       bool
	Header     bool
	LocalLogin bool
}

// SSO signs the owner in through an identity provider or a trusted proxy
// (input port). All methods end in the normal session creation.
type SSO interface {
	Status() SSOStatus
	// BeginOIDC returns the provider URL and the state to bind to the browser.
	BeginOIDC(ctx context.Context) (redirect, state string, err error)
	// CompleteOIDC finishes the code flow. With Require2FA it returns
	// domain.ErrSSOSecondFactorNeed and a ticket for CompleteSecondFactor.
	CompleteOIDC(ctx context.Context, state, code string, client domain.Client) (SessionResponse, string, error)
	// HeaderLogin trusts the user name a trusted proxy authenticated.
	HeaderLogin(ctx context.Context, user string, client domain.Client) (SessionResponse, string, error)
	CompleteSecondFactor(ctx context.Context, ticket, code string, client domain.Client) (SessionResponse, error)
}

type pendingSSO struct {
	nonce, verifier string
	expires         time.Time
}

type sso struct {
	deps Deps
	cfg  SSOConfig

	mu      sync.Mutex
	pending map[string]pendingSSO // state -> flow data
	tickets map[string]time.Time  // second-factor tickets -> expiry
}

// NewSSO returns the single sign-on use cases.
func NewSSO(deps Deps, cfg SSOConfig) SSO {
	return &sso{deps: deps, cfg: cfg, pending: map[string]pendingSSO{}, tickets: map[string]time.Time{}}
}

// Status implements SSO.
func (uc *sso) Status() SSOStatus {
	return SSOStatus{OIDC: uc.cfg.OIDC != nil, Header: len(uc.cfg.HeaderUsers) > 0, LocalLogin: uc.cfg.LocalLogin}
}

func (uc *sso) token() (string, error) {
	t, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return t, nil
}

func (uc *sso) prune(now time.Time) {
	for k, p := range uc.pending {
		if now.After(p.expires) {
			delete(uc.pending, k)
		}
	}
	for k, exp := range uc.tickets {
		if now.After(exp) {
			delete(uc.tickets, k)
		}
	}
}

// BeginOIDC implements SSO.
func (uc *sso) BeginOIDC(ctx context.Context) (string, string, error) {
	if uc.cfg.OIDC == nil {
		return "", "", domain.ErrSSOUnavailable
	}
	state, err := uc.token()
	if err != nil {
		return "", "", err
	}
	nonce, err := uc.token()
	if err != nil {
		return "", "", err
	}
	verifier, err := uc.token()
	if err != nil {
		return "", "", err
	}
	url, err := uc.cfg.OIDC.AuthCodeURL(ctx, state, nonce, verifier)
	if err != nil {
		return "", "", kernel.NewPublicErrorf(kernel.ErrUnavailable, "The identity provider could not be reached: {detail}", "detail", err.Error())
	}
	now := uc.deps.Clock.Now()
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.prune(now)
	if len(uc.pending) >= maxPendingSSO {
		return "", "", domain.ErrTooManyAttempts
	}
	uc.pending[state] = pendingSSO{nonce: nonce, verifier: verifier, expires: now.Add(ssoStateTTL)}
	return url, state, nil
}

// CompleteOIDC implements SSO.
func (uc *sso) CompleteOIDC(ctx context.Context, state, code string, client domain.Client) (SessionResponse, string, error) {
	if uc.cfg.OIDC == nil {
		return SessionResponse{}, "", domain.ErrSSOUnavailable
	}
	now := uc.deps.Clock.Now()
	uc.mu.Lock()
	p, ok := uc.pending[state]
	delete(uc.pending, state) // single use
	uc.mu.Unlock()
	if !ok || state == "" || now.After(p.expires) {
		return SessionResponse{}, "", domain.ErrSSOStateInvalid
	}
	id, err := uc.cfg.OIDC.Exchange(ctx, code, p.verifier, p.nonce)
	if err != nil {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, uc.deps.Limiter.RecordFailure(client.IP, now))
		return SessionResponse{}, "", kernel.NewPublicErrorf(kernel.ErrUnauthorized, "The identity provider did not confirm the sign-in: {detail}", "detail", err.Error())
	}
	id.Provider = "oidc"
	return uc.finish(ctx, id, client)
}

// HeaderLogin implements SSO; the caller has checked that the request came
// from a trusted proxy.
func (uc *sso) HeaderLogin(ctx context.Context, user string, client domain.Client) (SessionResponse, string, error) {
	if len(uc.cfg.HeaderUsers) == 0 {
		return SessionResponse{}, "", domain.ErrSSOUnavailable
	}
	id := domain.ExternalIdentity{Provider: "header", Subject: user}
	if user == "" || !(domain.IdentityMapping{Subjects: uc.cfg.HeaderUsers}).Allows(id) {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, false)
		return SessionResponse{}, "", domain.ErrSSONotAllowed
	}
	return uc.finish(ctx, id, client)
}

func (uc *sso) finish(ctx context.Context, id domain.ExternalIdentity, client domain.Client) (SessionResponse, string, error) {
	if id.Provider == "oidc" && !uc.cfg.Mapping.Allows(id) {
		uc.deps.Events.record(ctx, domain.EventLoginFailed, domain.OwnerID, client, "sso: "+id.Subject+" not allowed")
		return SessionResponse{}, "", domain.ErrSSONotAllowed
	}
	owner, exists, err := loadAccount(ctx, uc.deps.Accounts)
	if err != nil {
		return SessionResponse{}, "", err
	}
	if !exists {
		return SessionResponse{}, "", domain.ErrSetupRequired
	}
	if uc.cfg.Require2FA && owner.TOTPEnabled() {
		ticket, err := uc.token()
		if err != nil {
			return SessionResponse{}, "", err
		}
		uc.mu.Lock()
		uc.prune(uc.deps.Clock.Now())
		uc.tickets[ticket] = uc.deps.Clock.Now().Add(ssoStateTTL)
		uc.mu.Unlock()
		return SessionResponse{}, ticket, domain.ErrSSOSecondFactorNeed
	}
	uc.deps.Limiter.Reset(client.IP)
	uc.deps.Events.loginSucceeded(ctx, domain.OwnerID, client, id.Provider)
	res, err := uc.deps.issuer().issue(ctx, domain.OwnerID, client)
	return res, "", err
}

// CompleteSecondFactor implements SSO.
func (uc *sso) CompleteSecondFactor(ctx context.Context, ticket, code string, client domain.Client) (SessionResponse, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(client.IP, now) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	uc.mu.Lock()
	exp, ok := uc.tickets[ticket]
	uc.mu.Unlock()
	if !ok || now.After(exp) {
		return SessionResponse{}, domain.ErrSSOStateInvalid
	}
	owner, err := uc.deps.Accounts.Get(ctx, domain.OwnerID)
	if err != nil {
		return SessionResponse{}, err
	}
	if err := uc.deps.checkSecondFactor(ctx, &owner, code, now); err != nil {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, uc.deps.Limiter.RecordFailure(client.IP, now))
		return SessionResponse{}, err
	}
	if err := uc.deps.Accounts.Update(ctx, owner); err != nil {
		return SessionResponse{}, fmt.Errorf("update account: %w", err)
	}
	uc.mu.Lock()
	delete(uc.tickets, ticket)
	uc.mu.Unlock()
	uc.deps.Limiter.Reset(client.IP)
	uc.deps.Events.loginSucceeded(ctx, domain.OwnerID, client, "sso+totp")
	return uc.deps.issuer().issue(ctx, domain.OwnerID, client)
}
