package usecase

import "time"

// Filter narrows ListBookmarks. Empty fields match everything; FolderID
// "none" selects unfiled bookmarks.
type Filter struct {
	FolderID string
	Tag      string
	Query    string
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
