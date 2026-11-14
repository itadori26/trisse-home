package gateway

import (
	"context"
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/platform/database/dbtest"
)

// Migration 0008 rebuilds auth_accounts for the kiosk role: owner, guests,
// their tabs and two-factor state must survive.
func TestKioskMigrationKeepsAccounts(t *testing.T) {
	ctx := context.Background()
	before := fstest.MapFS{}
	for _, n := range []string{"0001_init.sql", "0002_session_metadata.sql", "0003_two_factor.sql", "0004_security_events.sql", "0005_accounts.sql", "0006_api_tokens.sql", "0007_passkeys.sql"} {
		data, err := fs.ReadFile(Migrations(), n)
		if err != nil {
			t.Fatal(err)
		}
		before[n] = &fstest.MapFile{Data: data}
	}
	db := dbtest.Open(t, "auth", before)
	if _, err := db.ExecContext(ctx, `INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at, totp_secret_enc, disabled, tabs)
		VALUES ('owner', 'owner', 'owner', 'h1', '2026-11-01T00:00:00Z', '2026-11-01T00:00:00Z', 'secret', 0, '[]'),
		       ('g1', 'guest', 'anna', 'h2', '2026-11-02T00:00:00Z', '2026-11-02T00:00:00Z', '', 1, '["t1","t2"]')`); err != nil {
		t.Fatal(err)
	}
	m, err := fs.ReadFile(Migrations(), "0008_kiosk.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(m)); err != nil {
		t.Fatal(err)
	}
	var role, tabs, totp string
	var disabled int
	if err := db.QueryRowContext(ctx, `SELECT role, tabs, disabled FROM auth_accounts WHERE id = 'g1'`).Scan(&role, &tabs, &disabled); err != nil || role != "guest" || tabs != `["t1","t2"]` || disabled != 1 {
		t.Fatalf("guest: %s %s %d %v", role, tabs, disabled, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT totp_secret_enc FROM auth_accounts WHERE id = 'owner'`).Scan(&totp); err != nil || totp != "secret" {
		t.Fatalf("owner: %q %v", totp, err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at) VALUES ('o2', 'owner', 'x', 'h', '', '')`); err == nil {
		t.Fatal("still only one owner")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at) VALUES ('k1', 'kiosk', 'flur', '!kiosk', '', '')`); err != nil {
		t.Fatal("kiosk role allowed", err)
	}
}

func TestKioskRepository(t *testing.T) {
	ctx := context.Background()
	r := NewKioskRepository(dbtest.Open(t, "auth", Migrations()))
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	k := domain.Kiosk{AccountID: "k1", TokenHash: domain.HashToken("secret"), Options: domain.KioskOptions{RotateSeconds: 60, NightFrom: "22:00", NightTo: "06:00"}}
	if err := r.Save(ctx, k); err != nil {
		t.Fatal(err)
	}
	if got, err := r.FindByTokenHash(ctx, domain.HashToken("secret")); err != nil || got.Options != k.Options || !got.LastUsedAt.IsZero() {
		t.Fatalf("%+v %v", got, err)
	}
	k.TokenHash, k.LastUsedAt, k.LastUsedIP = domain.HashToken("new"), now, "192.168.1.5"
	_ = r.Save(ctx, k)
	if _, err := r.FindByTokenHash(ctx, domain.HashToken("secret")); !errors.Is(err, domain.ErrKioskNotFound) {
		t.Fatal("the old link must stop working", err)
	}
	if got, _ := r.Get(ctx, "k1"); !got.LastUsedAt.Equal(now) || got.LastUsedIP != "192.168.1.5" {
		t.Fatalf("%+v", got)
	}
	_ = r.Delete(ctx, "k1")
	if _, err := r.Get(ctx, "k1"); !errors.Is(err, domain.ErrKioskNotFound) {
		t.Fatal(err)
	}
}
