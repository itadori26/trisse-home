package usecase

import (
	"context"
	"testing"
)

func TestSeedSampleData(t *testing.T) {
	ctx := context.Background()
	repo := newMem()
	d := Deps{Repo: repo, Clock: clk{}, IDs: &ids{}}
	for range 2 {
		if err := NewSeedSampleData(d).Execute(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if len(repo.folders) != 2 || len(repo.bookmarks) != 6 {
		t.Fatalf("got %d folders, %d bookmarks", len(repo.folders), len(repo.bookmarks))
	}
}
