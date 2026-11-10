package usecase

import (
	"context"
	"testing"
)

func TestSeedSampleData(t *testing.T) {
	repo := &memRepo{}
	for range 2 {
		if err := NewSeedSampleData(repo).Execute(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := NewSearcher(repo).Resolve(context.Background(), "!maps Oldenburg")
	if len(repo.shortcuts) != 1 || got != "https://www.openstreetmap.org/search?query=Oldenburg" {
		t.Fatalf("unexpected %d %q", len(repo.shortcuts), got)
	}
}
