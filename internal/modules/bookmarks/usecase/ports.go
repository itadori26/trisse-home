package usecase

import (
	"context"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/kernel"
)

// Input ports – one per use case.

// ListBookmarks returns bookmarks matching a filter plus folders and tags.
type ListBookmarks interface {
	Execute(ctx context.Context, f Filter) (ListResponse, error)
}

// CreateBookmark stores a new bookmark.
type CreateBookmark interface {
	Execute(ctx context.Context, in Input) (BookmarkView, error)
}

// UpdateBookmark changes a bookmark.
type UpdateBookmark interface {
	Execute(ctx context.Context, id string, in Input) (BookmarkView, error)
}

// DeleteBookmark removes a bookmark.
type DeleteBookmark interface {
	Execute(ctx context.Context, id string) error
}

// CreateFolder adds a folder at the end.
type CreateFolder interface {
	Execute(ctx context.Context, name string) (FolderView, error)
}

// RenameFolder renames a folder.
type RenameFolder interface {
	Execute(ctx context.Context, id, name string) (FolderView, error)
}

// DeleteFolder removes a folder; its bookmarks become unfiled.
type DeleteFolder interface {
	Execute(ctx context.Context, id string) error
}

// Output ports.

// Repository persists bookmarks and folders.
type Repository interface {
	ListBookmarks(ctx context.Context) ([]domain.Bookmark, error)
	GetBookmark(ctx context.Context, id string) (domain.Bookmark, error)
	SaveBookmark(ctx context.Context, b domain.Bookmark) error
	DeleteBookmark(ctx context.Context, id string) error
	ListFolders(ctx context.Context) ([]domain.Folder, error)
	GetFolder(ctx context.Context, id string) (domain.Folder, error)
	SaveFolder(ctx context.Context, f domain.Folder) error
	// DeleteFolder removes the folder and unfiles its bookmarks atomically.
	DeleteFolder(ctx context.Context, id string) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo  Repository
	Clock kernel.Clock
	IDs   kernel.IDGenerator
}
