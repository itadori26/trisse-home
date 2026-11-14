package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestMailStore(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "parcels", Migrations())
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	m := NewMailStore(db, prefixCipher{}, func() time.Time { return now })
	s, st, err := m.LoadMail(ctx)
	if err != nil || s.Enabled || s.Port != 993 || s.Folder != "INBOX" || st.LastUID != 0 {
		t.Fatalf("defaults: %+v %+v %v", s, st, err)
	}
	s = domain.MailSettings{Enabled: true, Host: "imap.example.com", Port: 993, Username: "me", Password: "pw", Folder: "Pakete", Interval: 30 * time.Minute}
	if err := m.SaveMail(ctx, s, usecase.MailState{UIDValidity: 7, LastUID: 42, LastRun: now, Found: 3}); err != nil {
		t.Fatal(err)
	}
	var enc string
	_ = db.QueryRowContext(ctx, `SELECT password_enc FROM parcels_mail`).Scan(&enc)
	if enc != "enc:pw" {
		t.Fatalf("password stored as %q", enc)
	}
	got, st, _ := m.LoadMail(ctx)
	if got != s || st.LastUID != 42 || st.Found != 3 || !st.LastRun.Equal(now) {
		t.Fatalf("%+v %+v", got, st)
	}
	if d, _ := m.Dismissed(ctx, "X1"); d {
		t.Fatal("nothing dismissed yet")
	}
	_ = m.Dismiss(ctx, "X1")
	_ = m.Dismiss(ctx, "X1")
	if d, _ := m.Dismissed(ctx, "X1"); !d {
		t.Fatal("dismissed")
	}
}
