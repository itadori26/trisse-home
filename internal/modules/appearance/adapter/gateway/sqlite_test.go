package gateway

import (
	"context"
	"testing"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "appearance", Migrations()))
	if a, err := repo.Load(ctx); err != nil || a.Theme() != domain.ThemeSystem {
		t.Fatalf("default not returned: %v", err)
	}
	for _, theme := range []domain.Theme{domain.ThemeDark, domain.ThemeLight} {
		a, _ := domain.NewAppearance(theme, "#123456", "", "body{}")
		if err := repo.Save(ctx, a); err != nil {
			t.Fatal(err)
		}
		if got, _ := repo.Load(ctx); got.Theme() != theme || got.CustomCSS() != "body{}" {
			t.Fatalf("round trip failed for %s", theme)
		}
	}
}
