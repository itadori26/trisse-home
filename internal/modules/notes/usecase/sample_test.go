package usecase

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/notes/domain"
)

func TestSeedSampleData(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{m: map[string]domain.Note{}}
	d := Deps{Repo: repo, Clock: &clk{t: time.Unix(0, 0)}, IDs: &ids{}}
	for range 2 {
		if err := NewSeedSampleData(d).Execute(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if len(repo.m) != 1 {
		t.Fatalf("expected exactly one welcome note, got %d", len(repo.m))
	}
	for _, n := range repo.m {
		if n.Title() != "Welcome to trisse-home 👋" || !n.Pinned() {
			t.Fatalf("unexpected note %q", n.Title())
		}
	}
}
