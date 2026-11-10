package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/docker/domain"
)

type fakeSource struct {
	list []domain.Container
	err  error
}

func (f *fakeSource) Containers(context.Context) ([]domain.Container, error) { return f.list, f.err }

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func TestContainers(t *testing.T) {
	ctx := context.Background()
	src := &fakeSource{list: []domain.Container{{Name: "b", State: "running"}, {Name: "a", State: "exited"}}}
	uc := NewContainers(src, clk{})
	res, err := uc.List(ctx)
	if err != nil || res.Total != 2 || res.Stopped != 1 || res.Containers[0].Name != "a" || res.Error != "" {
		t.Fatalf("unexpected %+v %v", res, err)
	}
	src.err = errors.New("connection refused")
	if err := uc.Refresh(ctx); err == nil {
		t.Fatal("expected refresh error")
	}
	res, _ = uc.List(ctx)
	if res.Total != 2 || res.Error == "" {
		t.Fatalf("last known state not kept: %+v", res)
	}
}

func TestContainersUnavailableFromStart(t *testing.T) {
	uc := NewContainers(&fakeSource{err: errors.New("down")}, clk{})
	res, err := uc.List(context.Background())
	if err != nil || res.Error == "" || len(res.Containers) != 0 {
		t.Fatalf("unexpected %+v %v", res, err)
	}
}
