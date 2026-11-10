package domain

import (
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Field limits.
const (
	MaxTitleLength       = 200
	MaxDescriptionLength = 1000
	MaxURLLength         = 2048
	MaxTags              = 10
	MaxTagLength         = 30
	MaxFolderNameLength  = 60
)

// Domain errors of the bookmarks module.
var (
	ErrBookmarkNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The bookmark does not exist.")
	ErrFolderNotFound   = kernel.NewPublicError(kernel.ErrNotFound, "The folder does not exist.")
)

var tagPattern = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)

// Bookmark is a saved link.
type Bookmark struct {
	id          string
	folderID    string
	title       string
	url         string
	description string
	tags        []string
	createdAt   time.Time
}

// BookmarkInput carries the editable fields of a bookmark.
type BookmarkInput struct {
	FolderID    string
	Title       string
	URL         string
	Description string
	Tags        []string
}

// NewBookmark validates the input and creates a bookmark. An empty title
// falls back to the URL's host.
func NewBookmark(id string, in BookmarkInput, now time.Time) (Bookmark, error) {
	b := Bookmark{id: id, createdAt: now}
	if err := b.Update(in); err != nil {
		return Bookmark{}, err
	}
	return b, nil
}

// RestoreBookmark rebuilds a bookmark from persisted state.
func RestoreBookmark(id, folderID, title, rawURL, description string, tags []string, createdAt time.Time) Bookmark {
	return Bookmark{id: id, folderID: folderID, title: title, url: rawURL, description: description, tags: tags, createdAt: createdAt}
}

// Update replaces all editable fields after validation.
func (b *Bookmark) Update(in BookmarkInput) error {
	rawURL := strings.TrimSpace(in.URL)
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(rawURL) > MaxURLLength {
		return kernel.NewValidationError("url", "The URL must start with http:// or https://.")
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = u.Host
	}
	if utf8.RuneCountInString(title) > MaxTitleLength {
		return kernel.NewValidationError("title", "The title must be at most 200 characters long.")
	}
	desc := strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(desc) > MaxDescriptionLength {
		return kernel.NewValidationError("description", "The description must be at most 1000 characters long.")
	}
	tags, err := NormaliseTags(in.Tags)
	if err != nil {
		return err
	}
	b.folderID, b.title, b.url, b.description, b.tags = strings.TrimSpace(in.FolderID), title, rawURL, desc, tags
	return nil
}

// NormaliseTags lower-cases, trims, de-duplicates and validates tags.
func NormaliseTags(raw []string) ([]string, error) {
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		t = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(t), "#"))
		if t == "" || slices.Contains(out, t) {
			continue
		}
		if utf8.RuneCountInString(t) > MaxTagLength || !tagPattern.MatchString(t) {
			return nil, kernel.NewValidationError("tags", "Tags may only contain letters, digits, - and _ (max. 30 characters).")
		}
		out = append(out, t)
	}
	if len(out) > MaxTags {
		return nil, kernel.NewValidationError("tags", "A bookmark can have at most 10 tags.")
	}
	slices.Sort(out)
	return out, nil
}

// ID returns the identifier.
func (b Bookmark) ID() string { return b.id }

// FolderID returns the folder or "" if the bookmark is unfiled.
func (b Bookmark) FolderID() string { return b.folderID }

// Title returns the title.
func (b Bookmark) Title() string { return b.title }

// URL returns the link target.
func (b Bookmark) URL() string { return b.url }

// Description returns the optional description.
func (b Bookmark) Description() string { return b.description }

// Tags returns the sorted tags.
func (b Bookmark) Tags() []string { return slices.Clone(b.tags) }

// CreatedAt returns when the bookmark was created.
func (b Bookmark) CreatedAt() time.Time { return b.createdAt }

// MoveToFolder files the bookmark into folderID ("" = unfiled).
func (b *Bookmark) MoveToFolder(folderID string) { b.folderID = folderID }

// Folder groups bookmarks.
type Folder struct {
	id       string
	name     string
	position int
}

// NewFolder validates and creates a folder.
func NewFolder(id, name string, position int) (Folder, error) {
	f := Folder{id: id, position: position}
	if err := f.Rename(name); err != nil {
		return Folder{}, err
	}
	return f, nil
}

// RestoreFolder rebuilds a folder from persisted state.
func RestoreFolder(id, name string, position int) Folder {
	return Folder{id: id, name: name, position: position}
}

// Rename changes the folder name after validation.
func (f *Folder) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > MaxFolderNameLength {
		return kernel.NewValidationError("name", "The folder name must be between 1 and 60 characters long.")
	}
	f.name = name
	return nil
}

// ID returns the identifier.
func (f Folder) ID() string { return f.id }

// Name returns the display name.
func (f Folder) Name() string { return f.name }

// Position returns the sort position.
func (f Folder) Position() int { return f.position }
