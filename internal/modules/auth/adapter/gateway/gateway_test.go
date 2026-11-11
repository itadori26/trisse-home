package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestSQLiteRepositories(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "auth", Migrations())
	now := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	accounts := NewAccountRepository(db)
	if _, err := accounts.Get(ctx); !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	a, _ := domain.NewAccount("hash1", now)
	if err := accounts.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	if err := accounts.Create(ctx, a); !errors.Is(err, domain.ErrAlreadySetUp) {
		t.Fatalf("expected ErrAlreadySetUp, got %v", err)
	}
	_ = a.ChangePasswordHash("hash2", now.Add(time.Hour))
	if err := accounts.Update(ctx, a); err != nil {
		t.Fatal(err)
	}
	got, err := accounts.Get(ctx)
	if err != nil || got.PasswordHash() != "hash2" || !got.CreatedAt().Equal(now) {
		t.Fatalf("round trip failed: %+v %v", got, err)
	}

	sessions := NewSessionRepository(db)
	s, _ := domain.NewSession(domain.HashToken("raw"), "pub1", "csrf", domain.NewClient("Firefox", "192.168.1.5"), now, time.Hour)
	if err := sessions.Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	found, err := sessions.Find(ctx, s.TokenHash())
	if err != nil || found.CSRFToken() != "csrf" || !found.ExpiresAt().Equal(now.Add(time.Hour)) || found.Client().UserAgent != "Firefox" || found.PublicID() != "pub1" {
		t.Fatalf("find: %+v %v", found, err)
	}
	other, _ := domain.NewSession(domain.HashToken("raw2"), "pub2", "csrf2", domain.Client{}, now.Add(time.Minute), time.Hour)
	_ = sessions.Create(ctx, other)
	if active, _ := sessions.ListActive(ctx, now); len(active) != 2 || active[0].PublicID() != "pub2" {
		t.Fatalf("list active: %v", active)
	}
	if err := sessions.DeleteByPublicID(ctx, "nope"); !errors.Is(err, domain.ErrSessionNotListed) {
		t.Fatalf("expected not listed, got %v", err)
	}
	if n, err := sessions.DeleteAllExcept(ctx, s.TokenHash()); err != nil || n != 1 {
		t.Fatalf("delete all except: %d %v", n, err)
	}
	if n, err := sessions.DeleteExpired(ctx, now.Add(2*time.Hour)); err != nil || n != 1 {
		t.Fatalf("delete expired: %d %v", n, err)
	}
	if _, err := sessions.Find(ctx, s.TokenHash()); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestAttemptLimiter(t *testing.T) {
	l := NewAttemptLimiter(2, time.Minute)
	now := time.Now()
	l.RecordFailure("a", now)
	if !l.Allow("a", now) {
		t.Fatal("blocked too early")
	}
	l.RecordFailure("a", now)
	if l.Allow("a", now) {
		t.Fatal("not blocked after max failures")
	}
	if !l.Allow("b", now) {
		t.Fatal("other key blocked")
	}
	if !l.Allow("a", now.Add(time.Minute)) {
		t.Fatal("not unblocked after window")
	}
}

func TestSetupToken(t *testing.T) {
	if NewSetupToken("").Verify("") {
		t.Fatal("empty token must never verify")
	}
	if !NewSetupToken("abc").Verify("abc") || NewSetupToken("abc").Verify("abd") {
		t.Fatal("verification wrong")
	}
}
