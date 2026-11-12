// Package gateway stores followed artists in SQLite and reads MusicBrainz
// and the Cover Art Archive.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"trisse-home/internal/modules/releases/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// TransferTables are part of data exports; the follow list has no secrets.
var TransferTables = []sqlutil.Table{{Name: "releases_artists"}}

// ArtistRepository implements usecase.ArtistRepository.
type ArtistRepository struct{ db *sql.DB }

// NewArtistRepository returns a repository backed by db.
func NewArtistRepository(db *sql.DB) *ArtistRepository { return &ArtistRepository{db: db} }

// List returns the followed artists in the order they were added.
func (r *ArtistRepository) List(ctx context.Context) ([]domain.Artist, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM releases_artists ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("query artists: %w", err)
	}
	defer rows.Close()
	var out []domain.Artist
	for rows.Next() {
		var a domain.Artist
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return nil, fmt.Errorf("scan artist: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Add stores an artist (idempotent).
func (r *ArtistRepository) Add(ctx context.Context, a domain.Artist) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO releases_artists (id, name, position)
		VALUES (?, ?, (SELECT COALESCE(MAX(position), -1) + 1 FROM releases_artists)) ON CONFLICT (id) DO NOTHING`, a.ID, a.Name)
	if err != nil {
		return fmt.Errorf("insert artist: %w", err)
	}
	return nil
}

// Remove deletes an artist.
func (r *ArtistRepository) Remove(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM releases_artists WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete artist: %w", err)
	}
	return nil
}
