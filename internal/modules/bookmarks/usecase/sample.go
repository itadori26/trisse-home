package usecase

import (
	"context"
	"fmt"
)

// SeedSampleData creates example folders and bookmarks if there are none
// (input port).
type SeedSampleData interface {
	Execute(ctx context.Context) error
}

type seedSampleData struct{ d Deps }

// NewSeedSampleData returns the SeedSampleData use case.
func NewSeedSampleData(d Deps) SeedSampleData { return &seedSampleData{d: d} }

type sampleFolder struct {
	name      string
	bookmarks []Input
}

var sampleFolders = []sampleFolder{
	{"Self-hosting", []Input{
		{Title: "Docker documentation", URL: "https://docs.docker.com/", Tags: []string{"docker", "docs"}},
		{Title: "Nextcloud", URL: "https://nextcloud.com/", Tags: []string{"cloud"}},
		{Title: "AdGuard Home", URL: "https://github.com/AdguardTeam/AdGuardHome", Tags: []string{"dns"}},
		{Title: "awesome-selfhosted", URL: "https://awesome-selfhosted.net/", Tags: []string{"ideas"}},
	}},
	{"Everyday", []Input{
		{Title: "OpenStreetMap", URL: "https://www.openstreetmap.org/", Tags: []string{"maps"}},
		{Title: "Wikipedia", URL: "https://www.wikipedia.org/", Tags: []string{"reference"}},
	}},
}

// Execute implements SeedSampleData.
func (uc *seedSampleData) Execute(ctx context.Context) error {
	bookmarks, err := uc.d.Repo.ListBookmarks(ctx)
	if err != nil {
		return fmt.Errorf("list bookmarks: %w", err)
	}
	folders, err := uc.d.Repo.ListFolders(ctx)
	if err != nil {
		return fmt.Errorf("list folders: %w", err)
	}
	if len(bookmarks) > 0 || len(folders) > 0 {
		return nil
	}
	createFolder, createBookmark := NewCreateFolder(uc.d), NewCreateBookmark(uc.d)
	for _, sf := range sampleFolders {
		f, err := createFolder.Execute(ctx, sf.name)
		if err != nil {
			return err
		}
		for _, in := range sf.bookmarks {
			in.FolderID = f.ID
			if _, err := createBookmark.Execute(ctx, in); err != nil {
				return err
			}
		}
	}
	return nil
}
