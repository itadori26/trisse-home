package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/fuel/domain"
	"trisse-home/internal/platform/database/dbtest"
)

const key = "00000000-0000-0000-0000-000000000002"

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
	db := dbtest.Open(t, "fuel", Migrations())
	repo := NewSettingsRepository(db, prefixCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s := domain.Settings{APIKey: key, Lat: 52.5, Lon: 13.4, Place: "Berlin", RadiusKm: 5, Fuel: domain.FuelDiesel,
		Favourites: []domain.Favourite{{ID: "11111111-1111-1111-1111-111111111111", Name: "Aral"}}}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Load(ctx)
	if err != nil || got.APIKey != key || got.Fuel != domain.FuelDiesel || len(got.Favourites) != 1 || got.Favourites[0].Name != "Aral" {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

func TestTankerkoenig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apikey") != key {
			_, _ = w.Write([]byte(`{"ok":false,"message":"apikey nicht angegeben, falsch, oder im falschen Format"}`))
			return
		}
		switch r.URL.Path {
		case "/list.php":
			if r.URL.Query().Get("type") != "e10" || r.URL.Query().Get("sort") != "price" {
				t.Errorf("query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"ok":true,"stations":[{"id":"a","name":"JET","brand":"JET","street":"Hauptstr.","houseNumber":"1","place":"Berlin","dist":1.2,"price":1.739,"isOpen":true},
				{"id":"b","name":"Shell","brand":"Shell","street":"Weg","houseNumber":"","place":"Berlin","dist":2,"price":null,"isOpen":false}]}`))
		case "/prices.php":
			_, _ = w.Write([]byte(`{"ok":true,"prices":{"c":{"status":"open","e5":1.8,"e10":1.75,"diesel":false},"d":{"status":"closed"}}}`))
		}
	}))
	defer srv.Close()
	tk := NewTankerkoenig(srv.Client(), srv.URL+"/")
	s := domain.Settings{APIKey: key, Lat: 52.5, Lon: 13.4, RadiusKm: 5, Fuel: domain.FuelE10}
	list, err := tk.Around(context.Background(), s)
	if err != nil || len(list) != 2 || list[0].Address != "Hauptstr. 1, Berlin" || list[1].Price != 0 || list[1].Address != "Weg, Berlin" {
		t.Fatalf("around: %+v %v", list, err)
	}
	prices, err := tk.Prices(context.Background(), key, []string{"c", "d"})
	if err != nil || !prices["c"].Open || prices["c"].ByFuel[domain.FuelE10] != 1.75 || prices["c"].ByFuel[domain.FuelDiesel] != 0 || prices["d"].Open {
		t.Fatalf("prices: %+v %v", prices, err)
	}
	s.APIKey = "00000000-0000-0000-0000-000000000009"
	if _, err := tk.Around(context.Background(), s); err == nil || !strings.Contains(err.Error(), "apikey") {
		t.Fatalf("provider message: %v", err)
	}
}

func TestErrorsNeverContainTheKey(t *testing.T) {
	tk := NewTankerkoenig(&http.Client{}, "http://127.0.0.1:1")
	_, err := tk.Around(context.Background(), domain.Settings{APIKey: key, Fuel: domain.FuelE5, RadiusKm: 1})
	if err == nil || strings.Contains(err.Error(), key) {
		t.Fatalf("the URL with the key must not leak: %v", err)
	}
}
