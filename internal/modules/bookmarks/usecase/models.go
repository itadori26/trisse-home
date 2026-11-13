package usecase

import (
	"time"

	"trisse-home/internal/modules/bookmarks/domain"
)

// Filter narrows ListBookmarks. Empty fields match everything; FolderID
// "none" selects unfiled bookmarks.
type Filter struct {
	FolderID string
	Tag      string
	Query    string
	// Broken keeps only bookmarks whose last link check failed (#72).
	Broken bool
}

// UnfiledFolderID selects bookmarks without folder in a Filter.
const UnfiledFolderID = "none"

// Input carries the editable fields of a bookmark.
type Input struct {
	FolderID    string
	Title       string
	URL         string
	Description string
	Tags        []string
}

// BookmarkView is a bookmark at the use case boundary.
type BookmarkView struct {
	ID, FolderID, Title, URL, Description string
	Tags                                  []string
	CreatedAt                             time.Time
	// Link is the result of the last link check (#72).
	Link domain.LinkStatus
	// HasIcon reports whether a favicon is stored.
	HasIcon bool
}

// SettingsView are the outbound features of the module (#72).
type SettingsView struct {
	Enrich     bool
	CheckLinks bool
}

// FolderView is a folder with the number of bookmarks it contains.
type FolderView struct {
	ID       string
	Name     string
	Position int
	Count    int
}

// TagView is a tag with its usage count.
type TagView struct {
	Name  string
	Count int
}

// ListResponse is the output of ListBookmarks.
type ListResponse struct {
	Bookmarks []BookmarkView
	Folders   []FolderView
	Tags      []TagView
}
