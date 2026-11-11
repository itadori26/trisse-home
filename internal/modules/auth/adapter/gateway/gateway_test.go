package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestSQLiteRepositories(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "auth", Migrations())
	now := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	accounts := NewAccountRepository(db, prefixCipher{})
	if _, err := accounts.Get(ctx, domain.OwnerID); !errors.Is(err, domain.ErrAccountNotFound) {
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
	got, err := accounts.Get(ctx, domain.OwnerID)
	if err != nil || got.PasswordHash() != "hash2" || !got.CreatedAt().Equal(now) {
		t.Fatalf("round trip failed: %+v %v", got, err)
	}

	// Two-factor state round trip; the secret must be stored encrypted.
	_ = got.StartTOTPEnrollment("JBSWY3DPEHPK3PXP")
	_ = got.ConfirmTOTP(42, []string{"abcd-efgh"})
	if err := accounts.Update(ctx, got); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT totp_secret_enc FROM auth_accounts WHERE id = 'owner'`).Scan(&stored)
	if stored == "JBSWY3DPEHPK3PXP" || stored == "" {
		t.Fatalf("totp secret not encrypted: %q", stored)
	}
	withTOTP, err := accounts.Get(ctx, domain.OwnerID)
	if err != nil || !withTOTP.TOTPEnabled() || withTOTP.TwoFactor().LastStep != 42 || len(withTOTP.TwoFactor().RecoveryCodes) != 1 {
		t.Fatalf("totp round trip: %+v %v", withTOTP.TwoFactor(), err)
	}

	sessions := NewSessionRepository(db)
	s, _ := domain.NewSession(domain.OwnerID, domain.HashToken("raw"), "pub1", "csrf", domain.NewClient("Firefox", "192.168.1.5"), now, time.Hour)
	if err := sessions.Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	found, err := sessions.Find(ctx, s.TokenHash())
	if err != nil || found.CSRFToken() != "csrf" || !found.ExpiresAt().Equal(now.Add(time.Hour)) || found.Client().UserAgent != "Firefox" || found.PublicID() != "pub1" {
		t.Fatalf("find: %+v %v", found, err)
	}
	other, _ := domain.NewSession(domain.OwnerID, domain.HashToken("raw2"), "pub2", "csrf2", domain.Client{}, now.Add(time.Minute), time.Hour)
	_ = sessions.Create(ctx, other)
	if active, _ := sessions.ListActive(ctx, domain.OwnerID, now); len(active) != 2 || active[0].PublicID() != "pub2" {
		t.Fatalf("list active: %v", active)
	}
	if err := sessions.DeleteByPublicID(ctx, domain.OwnerID, "nope"); !errors.Is(err, domain.ErrSessionNotListed) {
		t.Fatalf("expected not listed, got %v", err)
	}
	if n, err := sessions.DeleteAllExcept(ctx, domain.OwnerID, s.TokenHash()); err != nil || n != 1 {
		t.Fatalf("delete all except: %d %v", n, err)
	}
	if n, err := sessions.DeleteExpired(ctx, now.Add(2*time.Hour)); err != nil || n != 1 {
		t.Fatalf("delete expired: %d %v", n, err)
	}
	if _, err := sessions.Find(ctx, s.TokenHash()); !errors.Is(err, domain.ErrSessionNotFound) {
		t.Fatalf("expected not found, got %v", err)
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

// prefixCipher is a reversible stand-in for AES-GCM.
type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestEventRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewEventRepository(dbtest.Open(t, "auth", Migrations()))
	old := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, at := range []time.Time{old, old.AddDate(0, 6, 0)} {
		e, _ := domain.NewSecurityEvent(fmt.Sprint(i), at, domain.EventLoginSucceeded, domain.OwnerID, domain.Client{UserAgent: "Firefox", IP: "10.0.0.2"}, "")
		if err := repo.Add(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.List(ctx, 10)
	if err != nil || len(list) != 2 || list[0].ID != "1" || list[0].Client.UserAgent != "Firefox" {
		t.Fatalf("%+v %v", list, err)
	}
	if known, _ := repo.KnownDevice(ctx, domain.OwnerID, "Firefox"); !known {
		t.Fatal("device must be known")
	}
	if known, _ := repo.KnownDevice(ctx, domain.OwnerID, "Safari"); known {
		t.Fatal("unknown device reported as known")
	}
	if n, _ := repo.DeleteBefore(ctx, old.AddDate(0, 1, 0)); n != 1 {
		t.Fatalf("pruned %d", n)
	}
}

func TestAttemptLimiterBackoff(t *testing.T) {
	policy := domain.LoginThrottle{MaxFailures: 2, Window: time.Minute, BaseLockout: time.Minute, MaxLockout: time.Hour}
	l := NewAttemptLimiter(policy)
	now := time.Unix(0, 0)
	if l.RecordFailure("a", now) || !l.Allow("a", now) {
		t.Fatal("one failure must not lock")
	}
	if !l.RecordFailure("a", now) || l.Allow("a", now.Add(59*time.Second)) {
		t.Fatal("second failure must lock for a minute")
	}
	now = now.Add(time.Minute)
	if !l.Allow("a", now) {
		t.Fatal("lock must expire")
	}
	l.RecordFailure("a", now)
	l.RecordFailure("a", now)
	if l.Allow("a", now.Add(119*time.Second)) || !l.Allow("a", now.Add(2*time.Minute)) {
		t.Fatal("second lockout must last twice as long")
	}
	l.Reset("a")
	if !l.Allow("a", now) {
		t.Fatal("reset must unlock")
	}
}

func TestGuestAccountsAndMigration(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "auth", Migrations())
	now := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)
	accounts := NewAccountRepository(db, prefixCipher{})
	owner, _ := domain.NewAccount("hash", now)
	_ = accounts.Create(ctx, owner)
	g, _ := domain.NewGuest("g1", "anna", "ghash", []string{"tab1"}, now)
	if err := accounts.Create(ctx, g); err != nil {
		t.Fatal(err)
	}
	dup, _ := domain.NewGuest("g2", "anna", "x", nil, now)
	if err := accounts.Create(ctx, dup); !errors.Is(err, domain.ErrGuestNameTaken) {
		t.Fatalf("duplicate name: %v", err)
	}
	got, err := accounts.FindByName(ctx, "anna")
	if err != nil || got.ID() != "g1" || got.IsOwner() || len(got.Tabs()) != 1 {
		t.Fatalf("%+v %v", got, err)
	}
	_ = got.SetGuestAccess([]string{"tab2", "tab3"}, true, now)
	_ = accounts.Update(ctx, got)
	again, _ := accounts.Get(ctx, "g1")
	if !again.Disabled() || len(again.Tabs()) != 2 {
		t.Fatalf("update: %+v", again)
	}
	if list, _ := accounts.List(ctx); len(list) != 2 || !list[0].IsOwner() {
		t.Fatalf("list: %v", list)
	}
	sessions := NewSessionRepository(db)
	gs, _ := domain.NewSession("g1", domain.HashToken("g"), "pg", "c", domain.Client{}, now, time.Hour)
	os, _ := domain.NewSession(domain.OwnerID, domain.HashToken("o"), "po", "c", domain.Client{}, now, time.Hour)
	_ = sessions.Create(ctx, gs)
	_ = sessions.Create(ctx, os)
	if l, _ := sessions.ListActive(ctx, "g1", now); len(l) != 1 || l[0].AccountID() != "g1" {
		t.Fatalf("guest sessions: %v", l)
	}
	_ = sessions.DeleteAllFor(ctx, "g1")
	if _, err := sessions.Find(ctx, os.TokenHash()); err != nil {
		t.Fatal("owner session must survive")
	}
	if err := accounts.Delete(ctx, "g1"); err != nil {
		t.Fatal(err)
	}
	if _, err := accounts.Get(ctx, "g1"); !errors.Is(err, domain.ErrGuestNotFound) {
		t.Fatalf("deleted guest: %v", err)
	}
}
