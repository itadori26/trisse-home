package gateway

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/shared/sqlutil"
)

// policyRow is the persisted JSON form of domain.Policy.
type policyRow struct {
	Severities map[string]string `json:"severities,omitempty"`
	Quiet      struct {
		Enabled bool   `json:"enabled"`
		Start   string `json:"start"`
		End     string `json:"end"`
		Days    []int  `json:"days"`
	} `json:"quiet"`
}

func policyFrom(p domain.Policy) policyRow {
	var r policyRow
	r.Severities = map[string]string{}
	for k, v := range p.Severities {
		r.Severities[k] = string(v)
	}
	r.Quiet.Enabled, r.Quiet.Start, r.Quiet.End = p.Quiet.Enabled, p.Quiet.Start, p.Quiet.End
	for _, d := range p.Quiet.Days {
		r.Quiet.Days = append(r.Quiet.Days, int(d))
	}
	return r
}

func (r policyRow) domain() domain.Policy {
	p := domain.Policy{Severities: map[string]domain.Severity{}}
	for k, v := range r.Severities {
		p.Severities[k] = domain.Severity(v)
	}
	p.Quiet = domain.QuietHours{Enabled: r.Quiet.Enabled, Start: r.Quiet.Start, End: r.Quiet.End}
	for _, d := range r.Quiet.Days {
		if d >= 0 && d <= 6 {
			p.Quiet.Days = append(p.Quiet.Days, time.Weekday(d))
		}
	}
	return p
}

// ChannelRepository implements usecase.ChannelRepository.
type ChannelRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewChannelRepository returns the channel store; secrets are encrypted.
func NewChannelRepository(db *sql.DB, cipher Cipher) *ChannelRepository {
	return &ChannelRepository{db: db, cipher: cipher}
}

// List implements usecase.ChannelRepository.
func (r *ChannelRepository) List(ctx context.Context) ([]domain.Channel, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, kind, name, enabled, config, secret_enc FROM notification_channels ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query channels: %w", err)
	}
	defer rows.Close()
	var out []domain.Channel
	for rows.Next() {
		var c domain.Channel
		var kind, config, secret string
		var enabled int
		if err := rows.Scan(&c.ID, &kind, &c.Name, &enabled, &config, &secret); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		var row channelRow
		if err := json.Unmarshal([]byte(config), &row); err != nil {
			return nil, fmt.Errorf("decode channel: %w", err)
		}
		row.apply(&c)
		c.Kind, c.Enabled = domain.ChannelKind(kind), enabled == 1
		if secret != "" {
			if c.Secret, err = r.cipher.Decrypt(secret); err != nil {
				return nil, fmt.Errorf("decrypt channel secret (was SECRET_KEY changed?): %w", err)
			}
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Save implements usecase.ChannelRepository (upsert).
func (r *ChannelRepository) Save(ctx context.Context, c domain.Channel) error {
	secret := ""
	if c.Secret != "" {
		var err error
		if secret, err = r.cipher.Encrypt(c.Secret); err != nil {
			return fmt.Errorf("encrypt channel secret: %w", err)
		}
	}
	config, err := json.Marshal(channelRowFrom(c))
	if err != nil {
		return fmt.Errorf("encode channel: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO notification_channels (id, kind, name, enabled, config, secret_enc) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, enabled = excluded.enabled, config = excluded.config, secret_enc = excluded.secret_enc`,
		c.ID, string(c.Kind), c.Name, sqlutil.BoolToInt(c.Enabled), string(config), secret)
	if err != nil {
		return fmt.Errorf("upsert channel: %w", err)
	}
	return nil
}

// Delete implements usecase.ChannelRepository.
func (r *ChannelRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM notification_channels WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}

// MaintenanceRepository implements usecase.MaintenanceRepository.
type MaintenanceRepository struct{ db *sql.DB }

// NewMaintenanceRepository returns the maintenance window store.
func NewMaintenanceRepository(db *sql.DB) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

// ListEndingAfter implements usecase.MaintenanceRepository.
func (r *MaintenanceRepository) ListEndingAfter(ctx context.Context, t time.Time) ([]domain.MaintenanceWindow, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, scope, starts, ends, note FROM notification_maintenance WHERE ends > ? ORDER BY starts`, sqlutil.FormatTime(t))
	if err != nil {
		return nil, fmt.Errorf("query maintenance: %w", err)
	}
	defer rows.Close()
	var out []domain.MaintenanceWindow
	for rows.Next() {
		var w domain.MaintenanceWindow
		var scope, start, end string
		if err := rows.Scan(&w.ID, &scope, &start, &end, &w.Note); err != nil {
			return nil, fmt.Errorf("scan maintenance: %w", err)
		}
		if err := json.Unmarshal([]byte(scope), &w.Scope); err != nil {
			return nil, err
		}
		if w.Start, err = sqlutil.ParseTime(start); err != nil {
			return nil, err
		}
		if w.End, err = sqlutil.ParseTime(end); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// Save implements usecase.MaintenanceRepository (upsert).
func (r *MaintenanceRepository) Save(ctx context.Context, w domain.MaintenanceWindow) error {
	scope, _ := json.Marshal(w.Scope) // strings: cannot fail
	_, err := r.db.ExecContext(ctx, `INSERT INTO notification_maintenance (id, scope, starts, ends, note) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET scope = excluded.scope, starts = excluded.starts, ends = excluded.ends, note = excluded.note`,
		w.ID, string(scope), sqlutil.FormatTime(w.Start), sqlutil.FormatTime(w.End), w.Note)
	if err != nil {
		return fmt.Errorf("upsert maintenance: %w", err)
	}
	return nil
}

// Delete implements usecase.MaintenanceRepository.
func (r *MaintenanceRepository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM notification_maintenance WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete maintenance: %w", err)
	}
	return nil
}

// PendingRepository implements usecase.PendingRepository.
type PendingRepository struct{ db *sql.DB }

// NewPendingRepository returns the store for notifications deferred by quiet hours.
func NewPendingRepository(db *sql.DB) *PendingRepository { return &PendingRepository{db: db} }

// maxPending bounds the digest queue.
const maxPending = 500

// Add implements usecase.PendingRepository.
func (r *PendingRepository) Add(ctx context.Context, m domain.Message, at time.Time) error {
	var n int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM notification_pending`).Scan(&n); err != nil {
		return fmt.Errorf("count pending: %w", err)
	}
	if n >= maxPending {
		return nil // the digest says "… and N more" anyway; drop the overflow
	}
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("encode message: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO notification_pending (at, message) VALUES (?, ?)`, sqlutil.FormatTime(at), string(b))
	return err
}

// Take implements usecase.PendingRepository.
func (r *PendingRepository) Take(ctx context.Context) ([]domain.Message, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	out, err := pendingMessages(ctx, tx)
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM notification_pending`); err != nil {
		return nil, err
	}
	return out, tx.Commit()
}

func pendingMessages(ctx context.Context, tx *sql.Tx) ([]domain.Message, error) {
	rows, err := tx.QueryContext(ctx, `SELECT message FROM notification_pending ORDER BY seq`)
	if err != nil {
		return nil, fmt.Errorf("query pending: %w", err)
	}
	defer rows.Close()
	var out []domain.Message
	for rows.Next() {
		var raw string
		var m domain.Message
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if json.Unmarshal([]byte(raw), &m) == nil {
			out = append(out, m)
		}
	}
	return out, rows.Err()
}

// channelRow is the persisted JSON form of a channel. It has no secret
// field, so the secret can only ever be stored encrypted in secret_enc.
type channelRow struct {
	MinSeverity string   `json:"minSeverity"`
	Types       []string `json:"types"`
	URL         string   `json:"url"`
	Template    string   `json:"template"`
	SMTPHost    string   `json:"smtpHost"`
	SMTPPort    int      `json:"smtpPort"`
	Username    string   `json:"username"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	TLS         string   `json:"tls"`
}

func channelRowFrom(c domain.Channel) channelRow {
	return channelRow{MinSeverity: string(c.MinSeverity), Types: c.Types, URL: c.URL, Template: c.Template,
		SMTPHost: c.SMTPHost, SMTPPort: c.SMTPPort, Username: c.Username, From: c.From, To: c.To, TLS: c.TLS}
}

func (r channelRow) apply(c *domain.Channel) {
	c.MinSeverity, c.Types, c.URL, c.Template = domain.Severity(r.MinSeverity), r.Types, r.URL, r.Template
	c.SMTPHost, c.SMTPPort, c.Username, c.From, c.To, c.TLS = r.SMTPHost, r.SMTPPort, r.Username, r.From, r.To, r.TLS
}
