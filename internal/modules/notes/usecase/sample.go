package usecase

import (
	"context"
	"fmt"
)

// WelcomeNote is the pinned Markdown note created on a fresh installation.
const WelcomeNote = `# Welcome to trisse-home 👋

This start page was filled with **sample data** so you can see what it can do.

- **Edit** (top right): drag widgets by their title bar, resize them at the
  bottom-right corner, add widgets and tabs.
- **⚙ on a widget** (edit mode): configure it, e.g. choose your weather location.
- **Settings**: appearance, search engine & shortcuts, services, feeds,
  bookmark folders, calendar (CalDAV) and AdGuard Home.
- **Search**: DuckDuckGo !bangs work — try ` + "`!w clean architecture`" + ` or the
  custom shortcut ` + "`!maps Berlin`" + `.
- The **Server** tab shows host metrics, Docker containers and service status.

Delete whatever you don't need — this note included.`

// SeedSampleData creates the welcome note if there are no notes (input port).
type SeedSampleData interface {
	Execute(ctx context.Context) error
}

type seedSampleData struct{ d Deps }

// NewSeedSampleData returns the SeedSampleData use case.
func NewSeedSampleData(d Deps) SeedSampleData { return &seedSampleData{d: d} }

// Execute implements SeedSampleData.
func (uc *seedSampleData) Execute(ctx context.Context) error {
	existing, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list notes: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}
	_, err = NewSaveNote(uc.d).Execute(ctx, "", NoteInput{Body: WelcomeNote, Pinned: true})
	return err
}
