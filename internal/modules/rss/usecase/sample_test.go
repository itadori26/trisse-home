package usecase

import (
	"context"
	"testing"
)

func TestSeedSampleData(t *testing.T) {
	repo := newMem()
	uc := NewSeedSampleData(Deps{Repo: repo, Fetcher: &fakeFetcher{}, Clock: clk{}, IDs: &ids{}})
	urls := []string{"https://a.example/feed", "https://b.example/feed"}
	if added, err := uc.Execute(context.Background(), urls); err != nil || !added || len(repo.feeds) != 2 {
		t.Fatalf("first seed: %v %v %d", added, err, len(repo.feeds))
	}
	if added, _ := uc.Execute(context.Background(), urls); added {
		t.Fatal("seeding must not run twice")
	}
}
