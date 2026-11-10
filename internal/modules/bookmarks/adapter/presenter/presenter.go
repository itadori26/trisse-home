// Package presenter maps bookmarks responses to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/bookmarks/usecase"
)

// BookmarkDTO is a bookmark in API responses.
type BookmarkDTO struct {
	ID          string    `json:"id"`
	FolderID    string    `json:"folderId"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
}

// FolderDTO is a folder with its bookmark count.
type FolderDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Count    int    `json:"count"`
}

// TagDTO is a tag with its usage count.
type TagDTO struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ListDTO is returned by GET /api/bookmarks.
type ListDTO struct {
	Bookmarks []BookmarkDTO `json:"bookmarks"`
	Folders   []FolderDTO   `json:"folders"`
	Tags      []TagDTO      `json:"tags"`
}

// Bookmark maps a BookmarkView.
func Bookmark(v usecase.BookmarkView) BookmarkDTO {
	tags := v.Tags
	if tags == nil {
		tags = []string{}
	}
	return BookmarkDTO{ID: v.ID, FolderID: v.FolderID, Title: v.Title, URL: v.URL, Description: v.Description, Tags: tags, CreatedAt: v.CreatedAt}
}

// Folder maps a FolderView.
func Folder(v usecase.FolderView) FolderDTO { return FolderDTO(v) }

// List maps a ListResponse.
func List(r usecase.ListResponse) ListDTO {
	out := ListDTO{Bookmarks: make([]BookmarkDTO, 0, len(r.Bookmarks)), Folders: make([]FolderDTO, 0, len(r.Folders)), Tags: make([]TagDTO, 0, len(r.Tags))}
	for _, b := range r.Bookmarks {
		out.Bookmarks = append(out.Bookmarks, Bookmark(b))
	}
	for _, f := range r.Folders {
		out.Folders = append(out.Folders, Folder(f))
	}
	for _, t := range r.Tags {
		out.Tags = append(out.Tags, TagDTO(t))
	}
	return out
}
