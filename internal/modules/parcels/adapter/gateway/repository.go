// Package gateway stores parcels in SQLite and queries carrier APIs.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/parcels/domain"
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

// Repository implements usecase.Repository.
type Repository struct {
	db     *sql.DB
	cipher Cipher
}

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB, cipher Cipher) *Repository { return &Repository{db: db, cipher: cipher} }

const columns = `id, tracking_number, carrier, label, status_code, status_description, status_location, status_at,
	estimated_delivery, checked_at, error, created_at`

func scan(sc interface{ Scan(...any) error }) (domain.Parcel, error) {
	var p domain.Parcel
	var carrier, code, created string
	var statusAt, eta, checked *string
	if err := sc.Scan(&p.ID, &p.TrackingNumber, &carrier, &p.Label, &code, &p.Status.Description, &p.Status.Location,
		&statusAt, &eta, &checked, &p.Error, &created); err != nil {
		return p, err
	}
	p.Carrier, p.Status.Code = domain.Carrier(carrier), domain.StatusCode(code)
	var err error
	if p.Status.At, err = sqlutil.ParseNullableTime(statusAt); err != nil {
		return p, err
	}
	if p.EstimatedDelivery, err = sqlutil.ParseNullableTime(eta); err != nil {
		return p, err
	}
	if p.CheckedAt, err = sqlutil.ParseNullableTime(checked); err != nil {
		return p, err
	}
	p.CreatedAt, err = sqlutil.ParseTime(created)
	return p, err
}

// List implements usecase.Repository.
func (r *Repository) List(ctx context.Context) ([]domain.Parcel, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM parcels ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("query parcels: %w", err)
	}
	defer rows.Close()
	var out []domain.Parcel
	for rows.Next() {
		p, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan parcel: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Get implements usecase.Repository.
func (r *Repository) Get(ctx context.Context, id string) (domain.Parcel, error) {
	p, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM parcels WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Parcel{}, domain.ErrParcelNotFound
	}
	if err != nil {
		return domain.Parcel{}, fmt.Errorf("select parcel: %w", err)
	}
	return p, nil
}

// Save implements usecase.Repository (upsert).
func (r *Repository) Save(ctx context.Context, p domain.Parcel) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO parcels (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET label = excluded.label, status_code = excluded.status_code,
			status_description = excluded.status_description, status_location = excluded.status_location,
			status_at = excluded.status_at, estimated_delivery = excluded.estimated_delivery,
			checked_at = excluded.checked_at, error = excluded.error`,
		p.ID, p.TrackingNumber, string(p.Carrier), p.Label, string(p.Status.Code), p.Status.Description, p.Status.Location,
		sqlutil.FormatNullableTime(p.Status.At), sqlutil.FormatNullableTime(p.EstimatedDelivery), sqlutil.FormatNullableTime(p.CheckedAt),
		p.Error, sqlutil.FormatTime(p.CreatedAt))
	if err != nil {
		return fmt.Errorf("upsert parcel: %w", err)
	}
	return nil
}

// Delete implements usecase.Repository.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM parcels WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete parcel: %w", err)
	}
	return nil
}

// LoadSettings implements usecase.Repository; defaults without a row.
func (r *Repository) LoadSettings(ctx context.Context) (domain.Settings, error) {
	s := domain.Settings{HideDeliveredDays: domain.DefaultHideDeliveredDays}
	var enc string
	err := r.db.QueryRowContext(ctx, `SELECT dhl_api_key_enc, hide_delivered_days FROM parcels_settings WHERE id = 1`).Scan(&enc, &s.HideDeliveredDays)
	if errors.Is(err, sql.ErrNoRows) {
		return s, nil
	}
	if err != nil {
		return s, fmt.Errorf("select settings: %w", err)
	}
	if enc != "" {
		if s.DHLAPIKey, err = r.cipher.Decrypt(enc); err != nil {
			return s, fmt.Errorf("decrypt API key (was SECRET_KEY changed?): %w", err)
		}
	}
	return s, nil
}

// SaveSettings implements usecase.Repository.
func (r *Repository) SaveSettings(ctx context.Context, s domain.Settings) error {
	enc := ""
	if s.DHLAPIKey != "" {
		var err error
		if enc, err = r.cipher.Encrypt(s.DHLAPIKey); err != nil {
			return fmt.Errorf("encrypt API key: %w", err)
		}
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO parcels_settings (id, dhl_api_key_enc, hide_delivered_days) VALUES (1, ?, ?)
		ON CONFLICT (id) DO UPDATE SET dhl_api_key_enc = excluded.dhl_api_key_enc, hide_delivered_days = excluded.hide_delivered_days`,
		enc, s.HideDeliveredDays)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}
	return nil
}
