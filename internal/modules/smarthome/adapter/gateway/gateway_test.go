package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/smarthome/domain"
	"trisse-home/internal/platform/database/dbtest"
)

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + strings.ToUpper(s), nil }
func (prefixCipher) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "enc:") {
		return "", errors.New("bad ciphertext")
	}
	return strings.ToLower(strings.TrimPrefix(s, "enc:")), nil
}

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "smarthome", Migrations())
	repo := NewSettingsRepository(db, prefixCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s := domain.Settings{URL: "http://ha:8123", Token: "long-lived-token"}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT token_encrypted FROM smarthome_settings`).Scan(&stored)
	if strings.Contains(stored, "long-lived-token") {
		t.Fatalf("token stored in plain text: %q", stored)
	}
	if got, err := repo.Load(ctx); err != nil || got != s {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

func TestHomeAssistant(t *testing.T) {
	var called []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/states":
			_, _ = w.Write([]byte(`[
 {"entity_id":"switch.coffee","state":"on","attributes":{"friendly_name":"Coffee machine"},"last_changed":"2026-11-12T10:00:00+00:00"},
 {"entity_id":"sensor.power","state":"42.5","attributes":{"unit_of_measurement":"W"},"last_changed":"2026-11-12T10:01:00Z"},
 {"entity_id":"broken","state":"x","attributes":{}}
]`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/services/switch/toggle":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			called = append(called, body["entity_id"])
			_, _ = w.Write([]byte(`[]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	ha := NewHomeAssistant(srv.Client())
	s := domain.Settings{URL: srv.URL, Token: "tok"}
	list, err := ha.States(context.Background(), s)
	if err != nil || len(list) != 2 {
		t.Fatalf("states without invalid IDs: %+v %v", list, err)
	}
	if list[0].Name != "Coffee machine" || list[1].Name != "sensor.power" || list[1].Unit != "W" || list[0].LastChanged.Hour() != 10 {
		t.Fatalf("mapping: %+v", list)
	}
	if err := ha.Call(context.Background(), s, domain.Service{Domain: "switch", Name: "toggle"}, "switch.coffee"); err != nil {
		t.Fatal(err)
	}
	if len(called) != 1 || called[0] != "switch.coffee" {
		t.Fatalf("service call: %v", called)
	}
	s.Token = "bad"
	if _, err := ha.States(context.Background(), s); err == nil || !strings.Contains(err.Error(), "token was rejected") {
		t.Fatalf("expected token error, got %v", err)
	}
}
