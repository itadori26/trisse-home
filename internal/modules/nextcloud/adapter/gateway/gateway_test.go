package gateway

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/nextcloud/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func fakeNextcloud(t *testing.T, serverinfo bool) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /status.php", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"installed":true,"maintenance":false,"versionstring":"30.0.2"}`)
	})
	mux.HandleFunc("GET /ocs/v1.php/cloud/users/anna", func(w http.ResponseWriter, r *http.Request) {
		if u, p, _ := r.BasicAuth(); u != "anna" || p != "app-pw" || r.Header.Get("OCS-APIRequest") != "true" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `{"ocs":{"meta":{"statuscode":100},"data":{"quota":{"free":750,"used":250,"total":1000,"relative":25,"quota":1000}}}}`)
	})
	mux.HandleFunc("GET /ocs/v2.php/apps/serverinfo/api/v1/info", func(w http.ResponseWriter, _ *http.Request) {
		if !serverinfo {
			http.NotFound(w, nil)
			return
		}
		_, _ = io.WriteString(w, `{"ocs":{"meta":{"statuscode":200},"data":{"nextcloud":{"system":{"freespace":5000,"apps":{"num_updates_available":2}},
			"storage":{"num_users":3,"num_files":1200}},"server":{"database":{"size":"4096"}},"activeUsers":{"last5minutes":1,"last24hours":2}}}}`)
	})
	return httptest.NewServer(mux)
}

func TestAPIStatus(t *testing.T) {
	ctx := context.Background()
	srv := fakeNextcloud(t, true)
	defer srv.Close()
	s := domain.Settings{URL: srv.URL, Username: "anna", AppPassword: "app-pw"}
	st, err := NewAPI(srv.Client()).Status(ctx, s)
	if err != nil || st.Version != "30.0.2" || st.Quota.Total != 1000 || st.Server == nil || st.Server.AppUpdates != 2 || st.Server.DatabaseSize != 4096 {
		t.Fatalf("status %+v %v", st, err)
	}
	s.AppPassword = "wrong"
	if _, err := NewAPI(srv.Client()).Status(ctx, s); err == nil || !strings.Contains(err.Error(), "wrong") {
		t.Fatalf("wrong password: %v", err)
	}

	noInfo := fakeNextcloud(t, false)
	defer noInfo.Close()
	st, err = NewAPI(noInfo.Client()).Status(ctx, domain.Settings{URL: noInfo.URL, Username: "anna", AppPassword: "app-pw"})
	if err != nil || st.Server != nil || !strings.Contains(st.ServerHint, "serverinfo") {
		t.Fatalf("missing serverinfo must degrade gracefully: %+v %v", st, err)
	}
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "nextcloud", Migrations())
	repo := NewSettingsRepository(db, prefixCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	_ = repo.Save(ctx, domain.Settings{URL: "https://c", Username: "anna", AppPassword: "pw"})
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT app_password_encrypted FROM nextcloud_settings`).Scan(&stored)
	if got, err := repo.Load(ctx); err != nil || got.AppPassword != "pw" || stored != "enc:pw" {
		t.Fatalf("round trip %+v %q %v", got, stored, err)
	}
}
