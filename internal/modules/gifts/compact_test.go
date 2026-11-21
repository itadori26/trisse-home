package gifts

import (
	"context"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/gifts/adapter/gateway"
	"waylight/internal/modules/gifts/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return strconv.Itoa(s.n) }

func TestCompactState(t *testing.T) {
	m := New(dbtest.Open(t, "gifts", gateway.Migrations()), fixedClock{time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)}, &seqIDs{})
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	for _, in := range []domain.Input{{Person: "Ben", Title: "Book"}, {Person: "Ben", Title: "Ball"}, {Person: "Mia", Title: "Kite", Bought: true}} {
		if _, err := m.uc.Create(anna, in); err != nil {
			t.Fatal(err)
		}
	}
	s, err := m.CompactState(anna)
	if err != nil || s.Title.String() != "2 gift ideas" || *s.Count != 2 {
		t.Fatalf("anna: %+v %v", s, err)
	}
	if s, _ := m.CompactState(ben); !s.Empty || s.Title.String() != "No open gift ideas" {
		t.Fatalf("ben sees anna's count: %+v", s)
	}
}
