package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/immich/domain"
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
	db := dbtest.Open(t, "immich", Migrations())
	repo := NewSettingsRepository(db, prefixCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s := domain.Settings{URL: "http://immich:2283", APIKey: "secret-key"}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT api_key_encrypted FROM immich_settings`).Scan(&stored)
	if strings.Contains(stored, "secret-key") {
		t.Fatalf("API key stored in plain text: %q", stored)
	}
	if got, err := repo.Load(ctx); err != nil || got != s {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

const memoryAsset = "3f2b6c1e-0a4d-4e5f-9a8b-1c2d3e4f5a6b"

func TestAPI(t *testing.T) {
	storageAllowed := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/server/version" {
			_, _ = w.Write([]byte(`{"major":1,"minor":120,"patch":2}`))
			return
		}
		if r.Header.Get("x-api-key") != "key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/assets/statistics":
			_, _ = w.Write([]byte(`{"images":1200,"videos":34,"total":1234}`))
		case "/api/server/storage":
			if !storageAllowed {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			_, _ = w.Write([]byte(`{"diskSizeRaw":1000,"diskUseRaw":250}`))
		case "/api/memories":
			_, _ = w.Write([]byte(`[{"type":"on_this_day","data":{"year":2023},"assets":[{"id":"` + memoryAsset + `","type":"IMAGE"},{"id":"` + memoryAsset + `","type":"VIDEO"}]},
				{"type":"on_this_day","data":{"year":2026},"assets":[{"id":"` + memoryAsset + `","type":"IMAGE"}]}]`))
		case "/api/assets/" + memoryAsset + "/thumbnail":
			w.Header().Set("Content-Type", "image/webp")
			_, _ = w.Write([]byte("webp"))
		}
	}))
	defer srv.Close()
	a := NewAPI(srv.Client())
	s := domain.Settings{URL: srv.URL, APIKey: "key"}
	ctx := context.Background()

	if v, err := a.Version(ctx, s); err != nil || v != "1.120.2" {
		t.Fatalf("version: %q %v", v, err)
	}
	stats, err := a.Stats(ctx, s)
	if err != nil || stats.Photos != 1200 || stats.Videos != 34 || stats.DiskTotal != 1000 {
		t.Fatalf("stats: %+v %v", stats, err)
	}
	storageAllowed = false
	if stats, err := a.Stats(ctx, s); err != nil || stats.DiskTotal != 0 || stats.Photos != 1200 {
		t.Fatalf("the disk is optional: %+v %v", stats, err)
	}
	mem, err := a.Memories(ctx, s, time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC))
	if err != nil || len(mem) != 1 || mem[0].YearsAgo != 3 || len(mem[0].AssetIDs) != 1 {
		t.Fatalf("memories: photos of past years only: %+v %v", mem, err)
	}
	if th, err := a.Thumbnail(ctx, s, memoryAsset); err != nil || th.ContentType != "image/webp" {
		t.Fatalf("thumbnail: %+v %v", th, err)
	}
	s.APIKey = "bad"
	if _, err := a.Stats(ctx, s); err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Fatalf("expected key error, got %v", err)
	}
}
