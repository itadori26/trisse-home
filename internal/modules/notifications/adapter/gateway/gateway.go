// Package gateway implements the notification output ports: the encrypted
// settings repository and the ntfy sender.
//
// Ring: Interface Adapters.
package gateway

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/shared/sqlutil"
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

// Repository stores the settings with an encrypted token.
type Repository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB, cipher Cipher) *Repository { return &Repository{db: db, cipher: cipher} }

// rulesRow is the persisted JSON form of domain.Rules.
type rulesRow struct {
	Services      bool    `json:"services"`
	Containers    bool    `json:"containers"`
	Disk          bool    `json:"disk"`
	DiskThreshold float64 `json:"diskThreshold"`
	TLS           bool    `json:"tls"`
	TLSDays       int     `json:"tlsDays"`
	Backups       bool    `json:"backups"`
	Security      bool    `json:"security"`
}

// Load implements usecase.Repository.
func (r *Repository) Load(ctx context.Context) (domain.Settings, error) {
	var s domain.Settings
	var enabled int
	var tokenEnc, rules string
	err := r.db.QueryRowContext(ctx, `SELECT enabled, server_url, topic, token_encrypted, rules FROM notification_settings WHERE id = 1`).
		Scan(&enabled, &s.ServerURL, &s.Topic, &tokenEnc, &rules)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultSettings(), nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("select notification settings: %w", err)
	}
	s.Enabled = enabled == 1
	if tokenEnc != "" {
		if s.Token, err = r.cipher.Decrypt(tokenEnc); err != nil {
			return domain.Settings{}, fmt.Errorf("decrypt ntfy token (was SECRET_KEY changed?): %w", err)
		}
	}
	var rr rulesRow
	if err := json.Unmarshal([]byte(rules), &rr); err != nil {
		return domain.Settings{}, fmt.Errorf("decode rules: %w", err)
	}
	s.Rules = domain.Rules(rr)
	return s, nil
}

// Save implements usecase.Repository.
func (r *Repository) Save(ctx context.Context, s domain.Settings) error {
	tokenEnc := ""
	if s.Token != "" {
		var err error
		if tokenEnc, err = r.cipher.Encrypt(s.Token); err != nil {
			return fmt.Errorf("encrypt ntfy token: %w", err)
		}
	}
	rules, err := json.Marshal(rulesRow(s.Rules))
	if err != nil {
		return fmt.Errorf("encode rules: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO notification_settings (id, enabled, server_url, topic, token_encrypted, rules)
		VALUES (1, ?, ?, ?, ?, ?) ON CONFLICT (id) DO UPDATE SET enabled = excluded.enabled, server_url = excluded.server_url,
		topic = excluded.topic, token_encrypted = excluded.token_encrypted, rules = excluded.rules`,
		sqlutil.BoolToInt(s.Enabled), s.ServerURL, s.Topic, tokenEnc, string(rules))
	if err != nil {
		return fmt.Errorf("upsert notification settings: %w", err)
	}
	return nil
}

// NtfySender publishes messages with the ntfy JSON API.
type NtfySender struct{ client *http.Client }

// NewNtfySender returns a sender using client.
func NewNtfySender(client *http.Client) *NtfySender { return &NtfySender{client: client} }

type ntfyMessage struct {
	Topic    string   `json:"topic"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Priority int      `json:"priority"`
	Tags     []string `json:"tags,omitempty"`
}

// Send implements usecase.Sender.
func (n *NtfySender) Send(ctx context.Context, s domain.Settings, m domain.Message) error {
	body, err := json.Marshal(ntfyMessage{Topic: s.Topic, Title: m.Title, Message: m.Body, Priority: m.Priority, Tags: m.Tags})
	if err != nil {
		return fmt.Errorf("encode message: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.ServerURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.Token)
	}
	res, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("deliver to ntfy: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("ntfy answered %s", res.Status)
	}
	return nil
}
