// Package gateway implements the AdGuard output ports: the encrypted
// settings repository and the AdGuard Home API client.
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
	"slices"
	"time"

	"trisse-home/internal/modules/adguard/domain"
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

// SettingsRepository stores the settings with an encrypted password.
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
	err := r.db.QueryRowContext(ctx, `SELECT url, username, password_encrypted FROM adguard_settings WHERE id = 1`).Scan(&s.URL, &s.Username, &enc)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if s.Password, err = r.cipher.Decrypt(enc); err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt password (was SECRET_KEY changed?): %w", err)
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.Password)
	if err != nil {
		return fmt.Errorf("encrypt password: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO adguard_settings (id, url, username, password_encrypted) VALUES (1, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, username = excluded.username, password_encrypted = excluded.password_encrypted`,
		s.URL, s.Username, enc)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

const (
	maxResponseBytes = 2 << 20
	topBlockedLimit  = 5
)

// API queries the AdGuard Home control API.
type API struct{ client *http.Client }

// NewAPI returns a client.
func NewAPI(client *http.Client) *API { return &API{client: client} }

type statsPayload struct {
	Queries           int              `json:"num_dns_queries"`
	Blocked           int              `json:"num_blocked_filtering"`
	AvgProcessingTime float64          `json:"avg_processing_time"` // seconds
	TopBlocked        []map[string]int `json:"top_blocked_domains"`
}

type statusPayload struct {
	ProtectionEnabled bool `json:"protection_enabled"`
}

// Stats implements usecase.StatsSource.
func (a *API) Stats(ctx context.Context, s domain.Settings) (domain.Stats, error) {
	var st statsPayload
	if err := a.get(ctx, s, "/control/stats", &st); err != nil {
		return domain.Stats{}, err
	}
	var status statusPayload
	if err := a.get(ctx, s, "/control/status", &status); err != nil {
		return domain.Stats{}, err
	}
	out := domain.Stats{
		Queries: st.Queries, Blocked: st.Blocked, ProtectionEnabled: status.ProtectionEnabled,
		AvgProcessingTime: time.Duration(st.AvgProcessingTime * float64(time.Second)),
	}
	for _, entry := range st.TopBlocked {
		for d, n := range entry {
			out.TopBlocked = append(out.TopBlocked, domain.BlockedDomain{Domain: d, Count: n})
		}
	}
	slices.SortFunc(out.TopBlocked, func(a, b domain.BlockedDomain) int { return b.Count - a.Count })
	if len(out.TopBlocked) > topBlockedLimit {
		out.TopBlocked = out.TopBlocked[:topBlockedLimit]
	}
	return out, nil
}

func (a *API) get(ctx context.Context, s domain.Settings, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(s.Username, s.Password)
	res, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("AdGuard Home answered %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from AdGuard Home: %w", err)
	}
	return nil
}
