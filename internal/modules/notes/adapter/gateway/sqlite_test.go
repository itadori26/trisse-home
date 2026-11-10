package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/notes/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepositoryOrdering(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "notes", Migrations()))
	old, _ := domain.NewNote("a", "old", "", false, time.Unix(1, 0))
	recent, _ := domain.NewNote("b", "recent", "", false, time.Unix(2, 0))
	pinned, _ := domain.NewNote("c", "pinned", "", true, time.Unix(0, 0))
	for _, n := range []domain.Note{old, recent, pinned} {
		if err := repo.Save(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 3 || list[0].ID() != "c" || list[1].ID() != "b" {
		t.Fatalf("unexpected order: %v", list)
	}
}
