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

// GetSettings returns the outbound features (#72).
type GetSettings interface {
	Execute(ctx context.Context) (SettingsView, error)
}

// SaveSettings switches the outbound features (#72).
type SaveSettings interface {
	Execute(ctx context.Context, s SettingsView) (SettingsView, error)
}

// EnrichBookmark fetches title (only if the title is still the host name)
// and favicon of a bookmark (#72).
type EnrichBookmark interface {
	Execute(ctx context.Context, id string) (BookmarkView, error)
}

// CheckLink checks one bookmark's link now (#72).
type CheckLink interface {
	Execute(ctx context.Context, id string) (BookmarkView, error)
}

// CheckLinks checks the links that are due (background job, #72); it
// returns how many were checked.
type CheckLinks interface {
	Execute(ctx context.Context) (int, error)
}

// GetIcon returns a stored favicon (#72).
type GetIcon interface {
	Execute(ctx context.Context, id string) (domain.Icon, error)
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

// SettingsRepository stores the module settings (#72).
type SettingsRepository interface {
	LoadSettings(ctx context.Context) (domain.Settings, error)
	SaveSettings(ctx context.Context, s domain.Settings) error
}

// LinkStatusRepository stores link check results (#72).
type LinkStatusRepository interface {
	ListLinkStatuses(ctx context.Context) (map[string]domain.LinkStatus, error)
	SaveLinkStatus(ctx context.Context, bookmarkID string, s domain.LinkStatus) error
}

// IconStore keeps favicons locally so the browser never contacts the
// bookmarked sites (#72).
type IconStore interface {
	SaveIcon(ctx context.Context, bookmarkID string, icon domain.Icon) error
	// Icon returns domain.ErrIconNotFound if none is stored.
	Icon(ctx context.Context, bookmarkID string) (domain.Icon, error)
	IconIDs(ctx context.Context) (map[string]bool, error)
}

// MetadataFetcher reads title and favicon of a page (#72).
type MetadataFetcher interface {
	Fetch(ctx context.Context, pageURL string) (domain.PageMetadata, error)
	FetchIcon(ctx context.Context, iconURL string) (domain.Icon, error)
}

// LinkChecker requests a URL and reports the HTTP status and final URL;
// failed is true when no response arrived (#72).
type LinkChecker interface {
	Check(ctx context.Context, rawURL string) (code int, finalURL string, failed bool)
}

// Deps bundles the output ports. Settings, Links, Icons, Metadata and
// Checker are optional: without them the outbound features stay off.
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Settings SettingsRepository
	Links    LinkStatusRepository
	Icons    IconStore
	Metadata MetadataFetcher
	Checker  LinkChecker
}
