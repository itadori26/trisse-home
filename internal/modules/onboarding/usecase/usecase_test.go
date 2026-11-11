package usecase

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/onboarding/domain"
)

type memRepo struct{ s domain.State }

func (m *memRepo) Load(context.Context) (domain.State, error)   { return m.s, nil }
func (m *memRepo) Save(_ context.Context, s domain.State) error { m.s = s; return nil }

type clk struct{}

func (clk) Now() time.Time { return time.Unix(42, 0) }

func TestOnboarding(t *testing.T) {
	ctx := context.Background()
	uc := NewOnboarding(&memRepo{}, clk{})
	if s, _ := uc.Status(ctx); s.Completed {
		t.Fatal("fresh state must be pending")
	}
	if s, _ := uc.Complete(ctx); !s.Completed || s.CompletedAt.Unix() != 42 {
		t.Fatalf("complete: %+v", s)
	}
	if s, _ := uc.Restart(ctx); s.Completed {
		t.Fatal("restart failed")
	}
}
