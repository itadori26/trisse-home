package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ a *domain.Appearance }

func (m *memRepo) Load(context.Context) (domain.Appearance, error) {
	if m.a == nil {
		return domain.Default(), nil
	}
	return *m.a, nil
}

func (m *memRepo) Save(_ context.Context, a domain.Appearance) error { m.a = &a; return nil }

func TestAppearanceUseCases(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	if got, _ := NewGetAppearance(repo).Execute(ctx); got.Theme != "system" {
		t.Fatalf("default theme = %q", got.Theme)
	}
	if _, err := NewUpdateAppearance(repo).Execute(ctx, UpdateAppearanceRequest{Theme: "neon"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("expected invalid, got %v", err)
	}
	if _, err := NewUpdateAppearance(repo).Execute(ctx, UpdateAppearanceRequest{Theme: "dark", AccentColor: "#112233"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := NewGetAppearance(repo).Execute(ctx); got.Theme != "dark" || got.AccentColor != "#112233" {
		t.Fatalf("not saved: %+v", got)
	}
}
