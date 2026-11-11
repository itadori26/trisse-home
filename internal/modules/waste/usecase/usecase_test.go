package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeSource struct{ err error }

func (f *fakeSource) Collections(context.Context, string, time.Time, time.Time) ([]domain.Collection, error) {
	d := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	return []domain.Collection{{Date: d, Summary: "Biotonne"}, {Date: d.AddDate(0, 0, 7), Summary: "Papier"}}, f.err
}

type clk struct{}

func (clk) Now() time.Time { return time.Date(2026, 11, 11, 19, 0, 0, 0, time.UTC) }

func TestWaste(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{err: errors.New("404")}
	uc := NewWaste(repo, src, clk{})
	if _, err := uc.Upcoming(ctx); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatalf("unconfigured: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, "https://example.org/a.ics", nil); err == nil || repo.s != nil {
		t.Fatal("broken subscription must not be saved")
	}
	src.err = nil
	if _, err := uc.SaveSettings(ctx, "https://example.org/a.ics", []domain.Rule{{Match: "bio", Label: "Organic"}}); err != nil {
		t.Fatal(err)
	}
	o, err := uc.Upcoming(ctx)
	if err != nil || len(o.Types) != 2 || o.Types[0].Style.Label != "Organic" || o.Types[0].DaysUntil != 1 {
		t.Fatalf("%+v %v", o, err)
	}
	src.err = errors.New("down")
	_ = uc.Refresh(ctx)
	if o, _ := uc.Upcoming(ctx); o.Error == "" || len(o.Types) != 2 {
		t.Fatalf("last known dates must stay: %+v", o)
	}
}
