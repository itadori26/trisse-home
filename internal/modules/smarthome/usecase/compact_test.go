package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/smarthome/domain"
)

func TestCompact(t *testing.T) {
	ctx := context.Background()
	uc, repo, _, _, _ := setup()
	repo.s = &domain.Settings{URL: "http://ha", Token: "token"}

	// An allow-listed switch comes with its toggle.
	s, e, err := uc.Compact(ctx, "", "switch.coffee")
	if err != nil || s.Title.String() != "Coffee" || s.Detail.String() != "Off" || e.Action != domain.ActionToggle {
		t.Fatalf("%+v %+v %v", s, e, err)
	}
	if s.UpdatedAt.IsZero() {
		t.Fatal("no update time")
	}

	// A switch outside the allow-list is shown without an action.
	if s, e, _ := uc.Compact(ctx, "", "switch.heater"); e.Action != domain.ActionNone || s.Detail.String() != "On" {
		t.Fatalf("heater %+v", s)
	}
	// A sensor shows its value with the unit and has no action.
	if s, e, _ := uc.Compact(ctx, "", "sensor.power"); e.Action != domain.ActionNone || s.Detail.String() != "42 W" {
		t.Fatalf("sensor %+v", s)
	}

	if _, _, err := uc.Compact(ctx, "", ""); err == nil {
		t.Fatal("a missing entity must be refused")
	}
	if _, _, err := uc.Compact(ctx, "", "switch.nope"); !errors.Is(err, domain.ErrEntityNotFound) {
		t.Fatalf("unknown entity: %v", err)
	}
	if _, _, err := uc.Compact(ctx, "", "../etc"); !errors.Is(err, domain.ErrEntityNotFound) {
		t.Fatalf("invalid entity: %v", err)
	}
}
