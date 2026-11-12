package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// TokenRepository persists API tokens (output port).
type TokenRepository interface {
	Create(ctx context.Context, t domain.APIToken) error
	List(ctx context.Context) ([]domain.APIToken, error)
	// FindByHash returns domain.ErrSessionInvalid for unknown hashes.
	FindByHash(ctx context.Context, h domain.TokenHash) (domain.APIToken, error)
	UpdateUse(ctx context.Context, t domain.APIToken) error
	// Delete returns domain.ErrTokenNotFound for unknown IDs.
	Delete(ctx context.Context, id string) error
}

// TokenView describes a token without its secret.
type TokenView struct {
	ID         string
	Name       string
	Scopes     []string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastUsedAt time.Time
	LastUsedIP string
}

// CreatedToken is returned once after creation; Token is never shown again.
type CreatedToken struct {
	TokenView
	Token string
}

// NewTokenRequest is the input of APITokens.Create; ExpiresInDays 0 = never.
type NewTokenRequest struct {
	Name          string
	Scopes        []string
	ExpiresInDays int
}

// APITokens manages personal API tokens of the owner (input port).
type APITokens interface {
	List(ctx context.Context) ([]TokenView, error)
	Create(ctx context.Context, req NewTokenRequest) (CreatedToken, error)
	Revoke(ctx context.Context, id string) error
	// Authenticate resolves a bearer token to a principal with its scopes.
	Authenticate(ctx context.Context, raw string, client domain.Client) (kernel.Principal, error)
}

type apiTokens struct {
	deps   Deps
	tokens TokenRepository
}

// NewAPITokens returns the API token use cases.
func NewAPITokens(deps Deps, tokens TokenRepository) APITokens {
	return &apiTokens{deps: deps, tokens: tokens}
}

func tokenView(t domain.APIToken) TokenView {
	return TokenView{ID: t.ID, Name: t.Name, Scopes: t.Scopes, CreatedAt: t.CreatedAt, ExpiresAt: t.ExpiresAt,
		LastUsedAt: t.LastUsedAt, LastUsedIP: t.LastUsedIP}
}

// List implements APITokens.
func (uc *apiTokens) List(ctx context.Context) ([]TokenView, error) {
	if err := ownerOnly(ctx); err != nil {
		return nil, err
	}
	list, err := uc.tokens.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	out := make([]TokenView, 0, len(list))
	for _, t := range list {
		out = append(out, tokenView(t))
	}
	return out, nil
}

// Create implements APITokens.
func (uc *apiTokens) Create(ctx context.Context, req NewTokenRequest) (CreatedToken, error) {
	existing, err := uc.List(ctx)
	if err != nil {
		return CreatedToken{}, err
	}
	if len(existing) >= domain.MaxAPITokens {
		return CreatedToken{}, domain.ErrTooManyTokens
	}
	if req.ExpiresInDays < 0 || req.ExpiresInDays > 3650 {
		return CreatedToken{}, kernel.NewValidationError("expiresInDays", "Choose between 0 (never) and 3650 days.")
	}
	secret, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return CreatedToken{}, fmt.Errorf("generate token: %w", err)
	}
	raw := domain.APITokenPrefix + secret
	now := uc.deps.Clock.Now()
	var expires time.Time
	if req.ExpiresInDays > 0 {
		expires = now.AddDate(0, 0, req.ExpiresInDays)
	}
	t, err := domain.NewAPIToken(uc.deps.IDs.NewID(), req.Name, domain.HashToken(raw), req.Scopes, now, expires)
	if err != nil {
		return CreatedToken{}, err
	}
	if err := uc.tokens.Create(ctx, t); err != nil {
		return CreatedToken{}, fmt.Errorf("store token: %w", err)
	}
	uc.deps.Events.record(ctx, domain.EventTokenCreated, domain.OwnerID, domain.Client{}, t.Name+" ("+strings.Join(t.Scopes, ", ")+")")
	return CreatedToken{TokenView: tokenView(t), Token: raw}, nil
}

// Revoke implements APITokens.
func (uc *apiTokens) Revoke(ctx context.Context, id string) error {
	if err := ownerOnly(ctx); err != nil {
		return err
	}
	if err := uc.tokens.Delete(ctx, id); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventTokenRevoked, domain.OwnerID, domain.Client{}, id)
	return nil
}

// Authenticate implements APITokens. Failures count towards the login
// lockout of the client, so tokens cannot be brute-forced.
func (uc *apiTokens) Authenticate(ctx context.Context, raw string, client domain.Client) (kernel.Principal, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(client.IP, now) {
		return kernel.Principal{}, domain.ErrTooManyAttempts
	}
	t, err := uc.tokens.FindByHash(ctx, domain.HashToken(raw))
	if err == nil && t.Expired(now) {
		err = domain.ErrSessionInvalid
	}
	if errors.Is(err, domain.ErrSessionInvalid) || !strings.HasPrefix(raw, domain.APITokenPrefix) {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, uc.deps.Limiter.RecordFailure(client.IP, now))
		return kernel.Principal{}, domain.ErrSessionInvalid
	}
	if err != nil {
		return kernel.Principal{}, fmt.Errorf("find token: %w", err)
	}
	if t.RecordUse(now, client.IP) {
		if err := uc.tokens.UpdateUse(ctx, t); err != nil {
			return kernel.Principal{}, fmt.Errorf("record token use: %w", err)
		}
		uc.deps.Events.record(ctx, domain.EventTokenUsed, domain.OwnerID, client, t.Name)
	}
	return kernel.Principal{AccountID: domain.OwnerID, Role: "token", Scopes: t.Scopes}, nil
}
