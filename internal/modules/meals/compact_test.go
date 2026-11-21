package meals

import (
	"context"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/meals/adapter/gateway"
	"waylight/internal/modules/meals/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return strconv.Itoa(s.n) }

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)
	m := New(dbtest.Open(t, "meals", gateway.Migrations()), fixedClock{now}, &seqIDs{}, time.UTC)
	ctx := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})

	s, err := m.CompactState(ctx)
	if err != nil || !s.Empty || s.Title.String() != "Nothing planned for today" {
		t.Fatalf("empty: %+v %v", s, err)
	}
	for _, in := range []domain.Input{
		{Date: "2026-11-23", Meal: "dinner", Dish: "Lasagne"},
		{Date: "2026-11-23", Meal: "lunch", Dish: "Soup"},
		{Date: "2026-11-24", Dish: "Tomorrow"},
	} {
		if _, err := m.uc.Create(ctx, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	s, err = m.CompactState(ctx)
	if err != nil || s.Empty || s.Title.String() != "Today: Lasagne" || s.Detail.String() != "and 1 more" {
		t.Fatalf("planned: %+v %v", s, err)
	}
}
