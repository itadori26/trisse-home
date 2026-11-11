package usecase

import (
	"context"
	"testing"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/shared/kernel"
)

func TestSeedSampleData(t *testing.T) {
	repo := &memRepo{}
	uc := NewSeedSampleData(Deps{Repo: repo, Status: &memStatus{m: map[string]domain.Status{}}, Observations: kernel.DiscardObservations{}, Clock: clk{}, IDs: &ids{}})
	if created, err := uc.Execute(context.Background()); err != nil || !created || len(repo.items) != 3 {
		t.Fatalf("first seed: %v %v %d", created, err, len(repo.items))
	}
	if created, _ := uc.Execute(context.Background()); created || len(repo.items) != 3 {
		t.Fatal("seeding must not run twice")
	}
}
