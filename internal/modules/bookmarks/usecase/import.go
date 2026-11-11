package usecase

import (
	"context"
	"fmt"
	"strings"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/kernel"
)

// ImportResult summarises a browser bookmark import.
type ImportResult struct {
	Created        int
	SkippedInvalid int
	SkippedExists  int
	FoldersCreated int
}

// ImportBookmarks imports bookmarks read from a browser export (input port).
type ImportBookmarks interface {
	Execute(ctx context.Context, items []domain.ImportCandidate) (ImportResult, error)
}

type importBookmarks struct{ d Deps }

// NewImportBookmarks returns the ImportBookmarks use case.
func NewImportBookmarks(d Deps) ImportBookmarks { return &importBookmarks{d: d} }

// Execute implements ImportBookmarks. Folders are matched by name
// (case-insensitive) or created; URLs that already exist are skipped.
func (uc *importBookmarks) Execute(ctx context.Context, items []domain.ImportCandidate) (ImportResult, error) {
	if len(items) > domain.MaxImport {
		return ImportResult{}, kernel.NewValidationError("file", fmt.Sprintf("At most %d bookmarks can be imported at once.", domain.MaxImport))
	}
	existing, err := uc.d.Repo.ListBookmarks(ctx)
	if err != nil {
		return ImportResult{}, fmt.Errorf("list bookmarks: %w", err)
	}
	seen := map[string]bool{}
	for _, b := range existing {
		seen[b.URL()] = true
	}
	folders, err := uc.d.Repo.ListFolders(ctx)
	if err != nil {
		return ImportResult{}, fmt.Errorf("list folders: %w", err)
	}
	folderIDs := map[string]string{}
	for _, f := range folders {
		folderIDs[strings.ToLower(f.Name())] = f.ID()
	}
	var res ImportResult
	now := uc.d.Clock.Now()
	for _, raw := range items {
		c := domain.SanitiseCandidate(raw)
		url := strings.TrimSpace(c.URL)
		if !domain.Importable(url) {
			res.SkippedInvalid++
			continue
		}
		if seen[url] {
			res.SkippedExists++
			continue
		}
		folderID := ""
		if c.FolderPath != "" {
			key := strings.ToLower(c.FolderPath)
			if id, ok := folderIDs[key]; ok {
				folderID = id
			} else {
				f, err := domain.NewFolder(uc.d.IDs.NewID(), c.FolderPath, len(folderIDs))
				if err != nil {
					res.SkippedInvalid++
					continue
				}
				if err := uc.d.Repo.SaveFolder(ctx, f); err != nil {
					return res, fmt.Errorf("save folder: %w", err)
				}
				folderIDs[key], folderID = f.ID(), f.ID()
				res.FoldersCreated++
			}
		}
		b, err := domain.NewBookmark(uc.d.IDs.NewID(), domain.BookmarkInput{
			FolderID: folderID, Title: c.Title, URL: url, Description: c.Description, Tags: c.Tags,
		}, now)
		if err != nil {
			res.SkippedInvalid++
			continue
		}
		if err := uc.d.Repo.SaveBookmark(ctx, b); err != nil {
			return res, fmt.Errorf("save bookmark: %w", err)
		}
		seen[url] = true
		res.Created++
	}
	return res, nil
}
