package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
	"trisse-home/internal/shared/sqlutil"
)

// MailStore implements usecase.MailStore; the password is encrypted.
type MailStore struct {
	db     *sql.DB
	cipher Cipher
	now    func() time.Time
}

// NewMailStore returns a store backed by db.
func NewMailStore(db *sql.DB, cipher Cipher, now func() time.Time) *MailStore {
	return &MailStore{db: db, cipher: cipher, now: now}
}

// LoadMail implements usecase.MailStore; defaults without a row.
func (m *MailStore) LoadMail(ctx context.Context) (domain.MailSettings, usecase.MailState, error) {
	s := domain.MailSettings{Port: domain.DefaultMailPort, Folder: "INBOX", Interval: domain.DefaultMailInterval}
	var st usecase.MailState
	var enc string
	var minutes int
	var lastRun *string
	err := m.db.QueryRowContext(ctx, `SELECT enabled, host, port, username, password_enc, folder, interval_minutes,
			uid_validity, last_uid, last_run, last_error, found FROM parcels_mail WHERE id = 1`).
		Scan(&s.Enabled, &s.Host, &s.Port, &s.Username, &enc, &s.Folder, &minutes, &st.UIDValidity, &st.LastUID, &lastRun, &st.LastError, &st.Found)
	if errors.Is(err, sql.ErrNoRows) {
		return s, st, nil
	}
	if err != nil {
		return s, st, fmt.Errorf("select mail settings: %w", err)
	}
	s.Interval = time.Duration(minutes) * time.Minute
	if st.LastRun, err = sqlutil.ParseNullableTime(lastRun); err != nil {
		return s, st, err
	}
	if enc != "" {
		if s.Password, err = m.cipher.Decrypt(enc); err != nil {
			return s, st, fmt.Errorf("decrypt mail password (was SECRET_KEY changed?): %w", err)
		}
	}
	return s, st, nil
}

// SaveMail implements usecase.MailStore.
func (m *MailStore) SaveMail(ctx context.Context, s domain.MailSettings, st usecase.MailState) error {
	enc := ""
	if s.Password != "" {
		var err error
		if enc, err = m.cipher.Encrypt(s.Password); err != nil {
			return fmt.Errorf("encrypt mail password: %w", err)
		}
	}
	_, err := m.db.ExecContext(ctx, `INSERT INTO parcels_mail (id, enabled, host, port, username, password_enc, folder, interval_minutes,
			uid_validity, last_uid, last_run, last_error, found) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET enabled = excluded.enabled, host = excluded.host, port = excluded.port,
			username = excluded.username, password_enc = excluded.password_enc, folder = excluded.folder,
			interval_minutes = excluded.interval_minutes, uid_validity = excluded.uid_validity, last_uid = excluded.last_uid,
			last_run = excluded.last_run, last_error = excluded.last_error, found = excluded.found`,
		s.Enabled, s.Host, s.Port, s.Username, enc, s.Folder, int(s.Interval/time.Minute),
		st.UIDValidity, st.LastUID, sqlutil.FormatNullableTime(st.LastRun), st.LastError, st.Found)
	if err != nil {
		return fmt.Errorf("upsert mail settings: %w", err)
	}
	return nil
}

// Dismissed implements usecase.MailStore.
func (m *MailStore) Dismissed(ctx context.Context, number string) (bool, error) {
	var n int
	if err := m.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM parcels_dismissed WHERE tracking_number = ?`, number).Scan(&n); err != nil {
		return false, fmt.Errorf("select dismissed: %w", err)
	}
	return n > 0, nil
}

// Dismiss implements usecase.MailStore.
func (m *MailStore) Dismiss(ctx context.Context, number string) error {
	_, err := m.db.ExecContext(ctx, `INSERT INTO parcels_dismissed (tracking_number, dismissed_at) VALUES (?, ?)
		ON CONFLICT (tracking_number) DO NOTHING`, number, sqlutil.FormatTime(m.now()))
	if err != nil {
		return fmt.Errorf("insert dismissed: %w", err)
	}
	return nil
}
