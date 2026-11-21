package chores

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/chores/adapter/gateway"
	"waylight/internal/modules/chores/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return strconv.Itoa(s.n) }

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	m := New(dbtest.Open(t, "chores", gateway.Migrations()), fixedClock{now}, &seqIDs{}, time.UTC)
	ctx := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})

	if s, _ := m.CompactState(ctx); !s.Empty || s.Title.String() != "No chores yet" {
		t.Fatalf("no chores: %+v", s)
	}
	if _, err := m.uc.Create(ctx, domain.Input{Title: "Vacuum", EveryDays: 7, Due: "2026-11-25"}, ""); err != nil {
		t.Fatal(err)
	}
	if s, _ := m.CompactState(ctx); !s.Empty || s.Level != kernel.CompactOK || s.Title.String() != "Nothing to do today" {
		t.Fatalf("nothing due: %+v", s)
	}
	for _, in := range []domain.Input{
		{Title: "Dishes", EveryDays: 1, Due: "2026-11-22", Assignees: []string{"Ben", "Anna"}},
		{Title: "Plants", EveryDays: 3, Due: "2026-11-23"},
	} {
		if _, err := m.uc.Create(ctx, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	s, err := m.CompactState(ctx)
	if err != nil || s.Title.String() != "2 tasks today" || *s.Count != 2 || s.Level != kernel.CompactAttention || s.Detail.String() != "Dishes (Ben)" {
		t.Fatalf("due: %+v %v", s, err)
	}
	if a := s.Action; a == nil || a.Method != http.MethodPost || a.Path != "/api/chores/2/done" || a.Label.String() != "Done: Dishes" {
		t.Fatalf("action: %+v", s.Action)
	}
}
