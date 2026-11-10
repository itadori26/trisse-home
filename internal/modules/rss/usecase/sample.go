package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/rss/domain"
)

// SeedSampleData subscribes to example feeds if there are none (input port).
// Feeds are stored without fetching; it reports whether feeds were added so
// the caller can refresh them.
type SeedSampleData interface {
	Execute(ctx context.Context, urls []string) (bool, error)
}

type seedSampleData struct{ d Deps }

// NewSeedSampleData returns the SeedSampleData use case.
func NewSeedSampleData(d Deps) SeedSampleData { return &seedSampleData{d: d} }

// Execute implements SeedSampleData.
func (uc *seedSampleData) Execute(ctx context.Context, urls []string) (bool, error) {
	existing, err := uc.d.Repo.ListFeeds(ctx)
	if err != nil {
		return false, fmt.Errorf("list feeds: %w", err)
	}
	if len(existing) > 0 || len(urls) == 0 {
		return false, nil
	}
	for _, u := range urls {
		f, err := domain.NewFeed(uc.d.IDs.NewID(), u, uc.d.Clock.Now())
		if err != nil {
			return false, err
		}
		if err := uc.d.Repo.CreateFeed(ctx, f); err != nil {
			return false, fmt.Errorf("create feed: %w", err)
		}
	}
	return true, nil
}
