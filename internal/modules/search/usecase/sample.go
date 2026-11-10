package usecase

import (
	"context"
	"fmt"
)

// SeedSampleData adds an example custom shortcut if there is none (input port).
type SeedSampleData interface {
	Execute(ctx context.Context) error
}

type seedSampleData struct{ repo Repository }

// NewSeedSampleData returns the SeedSampleData use case.
func NewSeedSampleData(repo Repository) SeedSampleData { return &seedSampleData{repo: repo} }

// Execute implements SeedSampleData.
func (uc *seedSampleData) Execute(ctx context.Context) error {
	existing, err := uc.repo.Shortcuts(ctx)
	if err != nil {
		return fmt.Errorf("load shortcuts: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}
	_, err = NewSearcher(uc.repo).AddShortcut(ctx, AddShortcutRequest{
		Keyword: "maps", Name: "OpenStreetMap", URLTemplate: "https://www.openstreetmap.org/search?query={query}",
	})
	return err
}
