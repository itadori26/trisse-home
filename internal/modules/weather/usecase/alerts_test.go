package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeWarnings struct {
	calls int
	err   error
}

func (f *fakeWarnings) Warnings(context.Context, domain.Location) ([]domain.Warning, error) {
	f.calls++
	return []domain.Warning{{ID: "w", Severity: domain.SeverityModerate, Expires: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)}}, f.err
}

type fakeRadar struct{ calls int }

func (f *fakeRadar) Radar(context.Context, domain.BBox, int, int) ([]byte, error) {
	f.calls++
	return []byte("png"), nil
}

type mclock struct{ t time.Time }

func (c *mclock) Now() time.Time { return c.t }

func TestAlerts(t *testing.T) {
	ctx := context.Background()
	w, r, c := &fakeWarnings{}, &fakeRadar{}, &mclock{t: time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)}
	uc := NewAlerts(w, r, c)
	if list, err := uc.Warnings(ctx, 48.86, 2.35); err != nil || len(list) != 0 || w.calls != 0 {
		t.Fatal("no DWD request outside Germany")
	}
	if list, err := uc.Warnings(ctx, 53.14, 8.21); err != nil || len(list) != 1 {
		t.Fatalf("%v %v", list, err)
	}
	_, _ = uc.Warnings(ctx, 53.141, 8.209) // rounds to the same location → cached
	if w.calls != 1 {
		t.Fatalf("cache not used: %d", w.calls)
	}
	c.t = c.t.Add(6 * time.Minute)
	w.err = errors.New("down")
	if list, err := uc.Warnings(ctx, 53.14, 8.21); err != nil || len(list) != 1 {
		t.Fatal("stale warnings must beat an error")
	}
	if _, err := uc.Radar(ctx, 48.86, 2.35); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("radar outside Germany")
	}
	_, _ = uc.Radar(ctx, 53.14, 8.21)
	img, _ := uc.Radar(ctx, 53.14, 8.21)
	if string(img.PNG) != "png" || r.calls != 1 {
		t.Fatalf("radar cache: %d", r.calls)
	}
}
