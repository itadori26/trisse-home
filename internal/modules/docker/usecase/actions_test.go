package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeController struct{ restarted []string }

func (f *fakeController) Restart(_ context.Context, name string) error {
	f.restarted = append(f.restarted, name)
	return nil
}

type fakeLogs struct{ lines int }

func (f *fakeLogs) Logs(_ context.Context, _ string, lines int) ([]string, error) {
	f.lines = lines
	return []string{"ready"}, nil
}

func TestActions(t *testing.T) {
	ctx := context.Background()
	ctrl, logs, refreshed := &fakeController{}, &fakeLogs{}, 0
	uc := NewActions(domain.ActionPolicy{Enabled: true, Allowed: []string{"homarr"}}, ctrl, logs,
		func(context.Context) error { refreshed++; return nil })
	if err := uc.Restart(ctx, "nextcloud-app"); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("restart outside allow-list: %v", err)
	}
	if err := uc.Restart(ctx, "homarr"); err != nil || len(ctrl.restarted) != 1 || refreshed != 1 {
		t.Fatalf("restart: %v %v %d", err, ctrl.restarted, refreshed)
	}
	if l, err := uc.Logs(ctx, "homarr"); err != nil || len(l) != 1 || logs.lines != domain.MaxLogLines {
		t.Fatalf("logs: %v %v", l, err)
	}
	if c := uc.Capabilities(); !c.Enabled || c.Containers[0] != "homarr" {
		t.Fatalf("capabilities %+v", c)
	}
}
