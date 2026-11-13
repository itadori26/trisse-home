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
	// Link is the result of the last link check (#72).
	Link LinkDTO `json:"link"`
	// HasIcon reports whether a favicon is stored (GET /api/bookmarks/{id}/icon).
	HasIcon bool `json:"hasIcon"`
}

// LinkDTO is a link check result: state "" (not checked), "ok",
// "redirect" or "broken"; checkedAt is empty when never checked.
type LinkDTO struct {
	State     string `json:"state"`
	Code      int    `json:"code"`
	Target    string `json:"target"`
	CheckedAt string `json:"checkedAt"`
}

// SettingsDTO switches the outbound features of the bookmarks module.
type SettingsDTO struct {
	Enrich     bool `json:"enrich"`
	CheckLinks bool `json:"checkLinks"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

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
	dto := BookmarkDTO{ID: v.ID, FolderID: v.FolderID, Title: v.Title, URL: v.URL, Description: v.Description, Tags: tags, CreatedAt: v.CreatedAt,
		HasIcon: v.HasIcon, Link: LinkDTO{State: string(v.Link.State), Code: v.Link.Code, Target: v.Link.Target}}
	if !v.Link.CheckedAt.IsZero() {
		dto.Link.CheckedAt = v.Link.CheckedAt.UTC().Format(time.RFC3339)
	}
	return dto
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
