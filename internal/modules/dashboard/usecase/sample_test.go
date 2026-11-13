package usecase

import (
	"testing"

	"trisse-home/internal/modules/dashboard/domain"
)

func TestSampleLayoutIsValid(t *testing.T) {
	for _, tab := range SampleLayout(`{}`) {
		var rects []domain.Rect
		for _, w := range tab.Widgets {
			r, err := domain.NewRect(w.X, w.Y, w.W, w.H)
			if err != nil {
				t.Fatalf("%s/%s: %v", tab.Name, w.Kind, err)
			}
			for _, o := range rects {
				if r.Overlaps(o) {
					t.Fatalf("%s/%s overlaps another widget", tab.Name, w.Kind)
				}
			}
			rects = append(rects, r)
		}
	}
}

func TestSeedSampleData(t *testing.T) {
	deps, repo := newDeps("clock", "weather")
	uc := NewSeedSampleData(deps, SampleLayout(`{"lat":1,"lon":2}`))
	if err := uc.Execute(ctx); err != nil {
		t.Fatal(err)
	}
	if len(repo.tabs) != 2 || len(repo.widgets) != 2 {
		t.Fatalf("expected 2 tabs and only the 2 available widgets, got %d/%d", len(repo.tabs), len(repo.widgets))
	}
	if err := uc.Execute(ctx); err != nil || len(repo.tabs) != 2 {
		t.Fatal("seeding must not run twice")
	}
}
