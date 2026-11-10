// Package gateway implements the bookmarks repository on SQLite.
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
	"strings"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Repository implements usecase.Repository.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// bookmarkRow is the persistence model of a bookmark.
type bookmarkRow struct {
	id, title, url, description, createdAt string
	folderID                               sql.NullString
	tags                                   sql.NullString
}

func (r bookmarkRow) toDomain() (domain.Bookmark, error) {
	created, err := sqlutil.ParseTime(r.createdAt)
	if err != nil {
		return domain.Bookmark{}, err
	}
	var tags []string
	if r.tags.Valid && r.tags.String != "" {
		tags = strings.Split(r.tags.String, ",")
	}
	return domain.RestoreBookmark(r.id, r.folderID.String, r.title, r.url, r.description, tags, created), nil
}

const bookmarkQuery = `SELECT b.id, b.folder_id, b.title, b.url, b.description, b.created_at,
	(SELECT group_concat(tag, ',') FROM (SELECT tag FROM bookmark_tags t WHERE t.bookmark_id = b.id ORDER BY tag))
	FROM bookmarks b`

func scanBookmark(s interface{ Scan(...any) error }) (domain.Bookmark, error) {
	var r bookmarkRow
	if err := s.Scan(&r.id, &r.folderID, &r.title, &r.url, &r.description, &r.createdAt, &r.tags); err != nil {
		return domain.Bookmark{}, err
	}
	return r.toDomain()
}

// ListBookmarks returns all bookmarks ordered by title.
func (r *Repository) ListBookmarks(ctx context.Context) ([]domain.Bookmark, error) {
	rows, err := r.db.QueryContext(ctx, bookmarkQuery+` ORDER BY b.title COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("query bookmarks: %w", err)
	}
	defer rows.Close()
	var out []domain.Bookmark
	for rows.Next() {
		b, err := scanBookmark(rows)
		if err != nil {
			return nil, fmt.Errorf("scan bookmark: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBookmark returns one bookmark.
func (r *Repository) GetBookmark(ctx context.Context, id string) (domain.Bookmark, error) {
	b, err := scanBookmark(r.db.QueryRowContext(ctx, bookmarkQuery+` WHERE b.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Bookmark{}, domain.ErrBookmarkNotFound
	}
	if err != nil {
		return domain.Bookmark{}, fmt.Errorf("select bookmark: %w", err)
	}
	return b, nil
}

// SaveBookmark inserts or updates a bookmark and replaces its tags.
func (r *Repository) SaveBookmark(ctx context.Context, b domain.Bookmark) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var folder any
	if b.FolderID() != "" {
		folder = b.FolderID()
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO bookmarks (id, folder_id, title, url, description, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET folder_id = excluded.folder_id, title = excluded.title,
			url = excluded.url, description = excluded.description`,
		b.ID(), folder, b.Title(), b.URL(), b.Description(), sqlutil.FormatTime(b.CreatedAt()))
	if err != nil {
		return fmt.Errorf("upsert bookmark: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM bookmark_tags WHERE bookmark_id = ?`, b.ID()); err != nil {
		return fmt.Errorf("clear tags: %w", err)
	}
	for _, t := range b.Tags() {
		if _, err := tx.ExecContext(ctx, `INSERT INTO bookmark_tags (bookmark_id, tag) VALUES (?, ?)`, b.ID(), t); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	return tx.Commit()
}

// DeleteBookmark removes a bookmark (tags cascade).
func (r *Repository) DeleteBookmark(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM bookmarks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}
	return nil
}

// ListFolders returns all folders ordered by position.
func (r *Repository) ListFolders(ctx context.Context) ([]domain.Folder, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, position FROM bookmark_folders ORDER BY position, name`)
	if err != nil {
		return nil, fmt.Errorf("query folders: %w", err)
	}
	defer rows.Close()
	var out []domain.Folder
	for rows.Next() {
		var id, name string
		var pos int
		if err := rows.Scan(&id, &name, &pos); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		out = append(out, domain.RestoreFolder(id, name, pos))
	}
	return out, rows.Err()
}

// GetFolder returns one folder.
func (r *Repository) GetFolder(ctx context.Context, id string) (domain.Folder, error) {
	var name string
	var pos int
	err := r.db.QueryRowContext(ctx, `SELECT name, position FROM bookmark_folders WHERE id = ?`, id).Scan(&name, &pos)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Folder{}, domain.ErrFolderNotFound
	}
	if err != nil {
		return domain.Folder{}, fmt.Errorf("select folder: %w", err)
	}
	return domain.RestoreFolder(id, name, pos), nil
}

// SaveFolder inserts or updates a folder.
func (r *Repository) SaveFolder(ctx context.Context, f domain.Folder) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO bookmark_folders (id, name, position) VALUES (?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET name = excluded.name, position = excluded.position`, f.ID(), f.Name(), f.Position())
	if err != nil {
		return fmt.Errorf("upsert folder: %w", err)
	}
	return nil
}

// DeleteFolder removes a folder; ON DELETE SET NULL unfiles its bookmarks.
func (r *Repository) DeleteFolder(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM bookmark_folders WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}
	return nil
}
