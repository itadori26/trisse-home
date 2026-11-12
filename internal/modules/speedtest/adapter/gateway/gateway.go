// Package gateway implements the speedtest output ports: the encrypted
// settings repository and the Speedtest Tracker API client.
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
	"sort"
	"time"

	"trisse-home/internal/modules/speedtest/domain"
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

// SettingsRepository stores the settings with an encrypted token.
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
	err := r.db.QueryRowContext(ctx, `SELECT url, token_encrypted, min_down_mbps, min_up_mbps FROM speedtest_settings WHERE id = 1`).
		Scan(&s.URL, &enc, &s.MinDownMbps, &s.MinUpMbps)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select settings: %w", err)
	}
	if s.Token, err = r.cipher.Decrypt(enc); err != nil {
		return domain.Settings{}, fmt.Errorf("decrypt token (was SECRET_KEY changed?): %w", err)
	}
	return s, nil
}

// Save implements usecase.SettingsRepository.
func (r *SettingsRepository) Save(ctx context.Context, s domain.Settings) error {
	enc, err := r.cipher.Encrypt(s.Token)
	if err != nil {
		return fmt.Errorf("encrypt token: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO speedtest_settings (id, url, token_encrypted, min_down_mbps, min_up_mbps) VALUES (1, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET url = excluded.url, token_encrypted = excluded.token_encrypted,
		min_down_mbps = excluded.min_down_mbps, min_up_mbps = excluded.min_up_mbps`,
		s.URL, enc, s.MinDownMbps, s.MinUpMbps)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}

const (
	maxResponseBytes = 4 << 20
	pageSize         = 500
)

// API reads results from the Speedtest Tracker API (v1).
type API struct{ client *http.Client }

// NewAPI returns a client.
func NewAPI(client *http.Client) *API { return &API{client: client} }

type result struct {
	Ping         float64 `json:"ping"`
	Download     float64 `json:"download"` // bytes per second
	Upload       float64 `json:"upload"`
	DownloadBits float64 `json:"download_bits"`
	UploadBits   float64 `json:"upload_bits"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

var timeLayouts = []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05"}

func parseTime(s string) (time.Time, bool) {
	for _, l := range timeLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Results implements usecase.SpeedtestSource: completed results since the
// given time, oldest first.
func (a *API) Results(ctx context.Context, s domain.Settings, since time.Time) ([]domain.Measurement, error) {
	q := url.Values{}
	q.Set("filter[start_at]", since.UTC().Format("2006-01-02"))
	q.Set("sort", "-created_at")
	q.Set("per.page", fmt.Sprint(pageSize))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+"/api/v1/results?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("Accept", "application/json")
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Speedtest Tracker answered %s", res.Status) //nolint:staticcheck // product name
	}
	var payload struct {
		Data []result `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("unexpected response from Speedtest Tracker: %w", err)
	}
	out := make([]domain.Measurement, 0, len(payload.Data))
	for _, r := range payload.Data {
		if r.Status != "" && r.Status != "completed" {
			continue // failed or running tests have no numbers
		}
		at, ok := parseTime(r.CreatedAt)
		if !ok || at.Before(since) {
			continue
		}
		down, up := r.DownloadBits, r.UploadBits
		if down == 0 {
			down = r.Download * 8
		}
		if up == 0 {
			up = r.Upload * 8
		}
		out = append(out, domain.Measurement{At: at, DownloadBps: down, UploadBps: up, PingMs: r.Ping})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}
