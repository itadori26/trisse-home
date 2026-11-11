// Package gateway stores the encrypted Nextcloud settings and queries the
// Nextcloud OCS APIs.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"

	"trisse-home/internal/modules/nextcloud/domain"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts secrets before they are written to the database.
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// SettingsRepository stores the settings with an encrypted app password.
type SettingsRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewSettingsRepository returns a repository backed by db.
func NewSettingsRepository(db *sql.DB, cipher Cipher) *SettingsRepository {
	return &SettingsRepository{db: db, cipher: cipher}
}

// Load implements usecase.SettingsRepository.
func (r *SettingsRepository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var enc string
	err := r.db.QueryRowContext(ctx, `SELECT url, username, app_password_encrypted FROM nextcloud_settings WHERE id = 1`).Scan(&s.URL, &s.Username, &enc)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if s.AppPassword, err = r.cipher.Decrypt(enc); err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt app password (was SECRET_KEY changed?): %w", err)
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.AppPassword)
	if err != nil {
		return fmt.Errorf("encrypt app password: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO nextcloud_settings (id, url, username, app_password_encrypted) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, username = excluded.username, app_password_encrypted = excluded.app_password_encrypted`,
		s.URL, s.Username, enc)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

const maxResponseBytes = 2 << 20

// OCS status codes signalling success (v1 and v2 API); they are not HTTP codes.
const (
	ocsV1OK = 100
	ocsV2OK = 200
)

// API queries status.php and the OCS APIs.
type API struct{ client *http.Client }

// NewAPI returns an API client.
func NewAPI(client *http.Client) *API { return &API{client: client} }

// httpStatusError carries the HTTP (or OCS) status of a failed request.
type httpStatusError struct{ code int }

func (e httpStatusError) Error() string { return fmt.Sprintf("HTTP %d", e.code) }

// get fetches path below the base URL and decodes JSON into v.
func (a *API) get(ctx context.Context, s domain.Settings, path string, auth bool, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Accept", "application/json")
	if auth {
		req.SetBasicAuth(s.Username, s.AppPassword)
	}
	res, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return httpStatusError{code: res.StatusCode}
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

type ocsMeta struct {
	StatusCode int `json:"statuscode"`
}

type userPayload struct {
	OCS struct {
		Meta ocsMeta `json:"meta"`
		Data struct {
			Quota struct {
				Used  int64   `json:"used"`
				Free  int64   `json:"free"`
				Quota float64 `json:"quota"`
			} `json:"quota"`
		} `json:"data"`
	} `json:"ocs"`
}

type serverInfoPayload struct {
	OCS struct {
		Meta ocsMeta `json:"meta"`
		Data struct {
			Nextcloud struct {
				System struct {
					FreeSpace int64 `json:"freespace"`
					Apps      struct {
						Updates int `json:"num_updates_available"`
					} `json:"apps"`
				} `json:"system"`
				Storage struct {
					Users int `json:"num_users"`
					Files int `json:"num_files"`
				} `json:"storage"`
			} `json:"nextcloud"`
			Server struct {
				Database struct {
					Size json.Number `json:"size"`
				} `json:"database"`
			} `json:"server"`
			ActiveUsers struct {
				Last5Minutes int `json:"last5minutes"`
				Last24Hours  int `json:"last24hours"`
			} `json:"activeUsers"`
		} `json:"data"`
	} `json:"ocs"`
}

// Status implements usecase.StatusSource. status.php and the user's quota
// are required; serverinfo is optional and explained via ServerHint.
func (a *API) Status(ctx context.Context, s domain.Settings) (domain.Status, error) {
	var st struct {
		VersionString string `json:"versionstring"`
		Maintenance   bool   `json:"maintenance"`
	}
	if err := a.get(ctx, s, "/status.php", false, &st); err != nil {
		return domain.Status{}, fmt.Errorf("status.php: %w", err)
	}
	out := domain.Status{Version: st.VersionString, Maintenance: st.Maintenance}

	var user userPayload
	if err := a.get(ctx, s, "/ocs/v1.php/cloud/users/"+url.PathEscape(s.Username)+"?format=json", true, &user); err != nil {
		var hs httpStatusError
		if errors.As(err, &hs) && hs.code == http.StatusUnauthorized {
			return domain.Status{}, errors.New("username or app password is wrong")
		}
		return domain.Status{}, fmt.Errorf("user api: %w", err)
	}
	if user.OCS.Meta.StatusCode != ocsV1OK {
		return domain.Status{}, fmt.Errorf("user api returned OCS status %d", user.OCS.Meta.StatusCode)
	}
	q := user.OCS.Data.Quota
	out.Quota = domain.Quota{Used: q.Used, Free: q.Free}
	if q.Quota > 0 {
		out.Quota.Total = int64(q.Quota)
	}

	var info serverInfoPayload
	err := a.get(ctx, s, "/ocs/v2.php/apps/serverinfo/api/v1/info?format=json&skipUpdate=false", true, &info)
	var hs httpStatusError
	switch {
	case errors.As(err, &hs):
		out.ServerHint = domain.ServerInfoHint(hs.code)
	case err != nil:
		out.ServerHint = domain.ServerInfoHint(0)
	case info.OCS.Meta.StatusCode != ocsV2OK && info.OCS.Meta.StatusCode != ocsV1OK:
		out.ServerHint = domain.ServerInfoHint(info.OCS.Meta.StatusCode)
	default:
		d := info.OCS.Data
		dbSize, _ := d.Server.Database.Size.Int64()
		out.Server = &domain.ServerInfo{
			FreeSpace: d.Nextcloud.System.FreeSpace, Users: d.Nextcloud.Storage.Users, Files: d.Nextcloud.Storage.Files,
			AppUpdates: d.Nextcloud.System.Apps.Updates, ActiveUsers5m: d.ActiveUsers.Last5Minutes, ActiveUsers1d: d.ActiveUsers.Last24Hours,
			DatabaseSize: dbSize,
		}
	}
	return out, nil
}
