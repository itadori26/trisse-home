package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/platform/database/dbtest"
)

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "notifications", Migrations())
	repo := NewRepository(db, prefixCipher{})
	if s, err := repo.Load(ctx); err != nil || s.Enabled || s.ServerURL != domain.DefaultServerURL {
		t.Fatalf("defaults %+v %v", s, err)
	}
	rules := domain.DefaultRules()
	rules.DiskThreshold = 80
	s, _ := domain.NewSettings(true, "https://ntfy.example", "home", "tk", rules)
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT token_encrypted FROM notification_settings`).Scan(&stored)
	if stored != "enc:tk" {
		t.Fatalf("token not encrypted: %q", stored)
	}
	got, err := repo.Load(ctx)
	if err != nil || got.Token != "tk" || got.Rules.DiskThreshold != 80 || !got.Enabled {
		t.Fatalf("round trip %+v %v", got, err)
	}
}

func TestNtfySender(t *testing.T) {
	var received ntfyMessage
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&received)
	}))
	defer srv.Close()
	s := domain.Settings{ServerURL: srv.URL, Topic: "home", Token: "tk"}
	err := NewNtfySender(srv.Client()).Send(context.Background(), s, domain.Message{Title: "Nextcloud ist offline ⚠", Body: "timeout", Priority: 4})
	if err != nil || received.Topic != "home" || received.Title != "Nextcloud ist offline ⚠" || auth != "Bearer tk" {
		t.Fatalf("unexpected %+v %q %v", received, auth, err)
	}
	s.ServerURL = srv.URL + "/missing"
	bad := httptest.NewServer(http.NotFoundHandler())
	defer bad.Close()
	s.ServerURL = bad.URL
	if err := NewNtfySender(bad.Client()).Send(context.Background(), s, domain.Message{}); err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestChannelSecretIsEncrypted(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "notifications", Migrations())
	repo := NewChannelRepository(db, prefixCipher{})
	c := domain.Channel{ID: "1", Kind: domain.ChannelGotify, Name: "G", Enabled: true, URL: "https://g.example", Secret: "app-token", MinSeverity: domain.SeverityWarning, Types: []string{"service"}}
	if err := repo.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
	var config, secret string
	_ = db.QueryRowContext(ctx, `SELECT config, secret_enc FROM notification_channels`).Scan(&config, &secret)
	if strings.Contains(config, "app-token") || secret == "" || secret == "app-token" {
		t.Fatalf("secret stored in plain text: %s / %s", config, secret)
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 1 || list[0].Secret != "app-token" || list[0].URL != c.URL || list[0].Types[0] != "service" {
		t.Fatalf("%+v %v", list, err)
	}
}
