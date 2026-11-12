package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

type memTokens struct{ list []domain.APIToken }

func (m *memTokens) Create(_ context.Context, t domain.APIToken) error {
	m.list = append(m.list, t)
	return nil
}
func (m *memTokens) List(context.Context) ([]domain.APIToken, error) { return m.list, nil }
func (m *memTokens) FindByHash(_ context.Context, h domain.TokenHash) (domain.APIToken, error) {
	for _, t := range m.list {
		if t.Hash == h {
			return t, nil
		}
	}
	return domain.APIToken{}, domain.ErrSessionInvalid
}
func (m *memTokens) UpdateUse(_ context.Context, t domain.APIToken) error {
	for i := range m.list {
		if m.list[i].ID == t.ID {
			m.list[i] = t
		}
	}
	return nil
}
func (m *memTokens) Delete(_ context.Context, id string) error {
	for i, t := range m.list {
		if t.ID == id {
			m.list = append(m.list[:i], m.list[i+1:]...)
			return nil
		}
	}
	return domain.ErrTokenNotFound
}

func TestAPITokens(t *testing.T) {
	e := newEnv().withAccount()
	repo := &memTokens{}
	events, _ := withEvents(e)
	uc := NewAPITokens(e.deps, repo)
	created, err := uc.Create(ctx, NewTokenRequest{Name: "Shortcuts", Scopes: []string{"todos:write"}, ExpiresInDays: 7})
	if err != nil || !strings.HasPrefix(created.Token, domain.APITokenPrefix) || created.ExpiresAt.IsZero() {
		t.Fatalf("%+v %v", created, err)
	}
	if strings.Contains(string(repo.list[0].Hash), created.Token) {
		t.Fatal("the raw token must not be stored")
	}
	client := domain.Client{IP: "10.0.0.5"}
	p, err := uc.Authenticate(ctx, created.Token, client)
	if err != nil || !p.IsToken() || p.Scopes[0] != "todos:write" {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err := uc.Authenticate(ctx, "th_wrong", client); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("wrong token: %v", err)
	}
	e.clock.now = e.clock.now.Add(8 * 24 * time.Hour)
	if _, err := uc.Authenticate(ctx, created.Token, domain.Client{IP: "10.0.0.6"}); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("expired token accepted: %v", err)
	}
	guest := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "g", Role: "guest"})
	if _, err := uc.List(guest); !errors.Is(err, domain.ErrOwnerOnly) {
		t.Fatalf("guest listed tokens: %v", err)
	}
	if err := uc.Revoke(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	seen := map[domain.EventType]bool{}
	for _, ty := range events.types() {
		seen[ty] = true
	}
	if !seen[domain.EventTokenCreated] || !seen[domain.EventTokenUsed] || !seen[domain.EventTokenRevoked] || !seen[domain.EventLoginFailed] {
		t.Fatalf("events: %v", events.types())
	}
}
