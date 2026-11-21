package usecase

import (
	"testing"

	"waylight/internal/modules/dashboard/domain"
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
	for _, tab := range repo.tabs {
		if !tab.Sample() {
			t.Fatalf("%s is not marked as an example tab", tab.Name())
		}
	}
	if err := uc.Execute(ctx); err != nil || len(repo.tabs) != 2 {
		t.Fatal("seeding must not run twice")
	}
}

// The demo showcase is the demo itself, so its tabs are not example tabs
// the template choice could replace (#670).
func TestSeedShowcaseIsNoSample(t *testing.T) {
	deps, repo := newDeps("clock")
	if err := NewSeedShowcase(deps, SampleLayout(`{}`)).Execute(ctx); err != nil {
		t.Fatal(err)
	}
	for _, tab := range repo.tabs {
		if tab.Sample() {
			t.Fatalf("%s is marked as an example tab", tab.Name())
		}
	}
}

// TestSampleTabs: "start empty" and a template remove the example tabs and
// close the gaps in the order; keeping them makes them ordinary tabs (#670).
func TestSampleTabs(t *testing.T) {
	deps, repo := newDeps("clock")
	if err := NewSeedSampleData(deps, SampleLayout(`{}`)).Execute(ctx); err != nil {
		t.Fatal(err)
	}
	own, _ := domain.NewTab("own", "Family", 2)
	_ = repo.CreateTab(ctx, own)
	uc := NewSampleTabs(deps)
	if err := uc.Discard(ctx); err != nil {
		t.Fatal(err)
	}
	if len(repo.tabs) != 1 || repo.tabs[0].ID() != "own" || repo.tabs[0].Position() != 0 {
		t.Fatalf("after discard: %+v", repo.tabs)
	}
	if len(repo.widgets) != 0 {
		t.Fatalf("widgets of the example tabs left: %d", len(repo.widgets))
	}

	deps, repo = newDeps("clock")
	if err := NewSeedSampleData(deps, SampleLayout(`{}`)).Execute(ctx); err != nil {
		t.Fatal(err)
	}
	uc = NewSampleTabs(deps)
	if err := uc.Keep(ctx); err != nil {
		t.Fatal(err)
	}
	if len(repo.tabs) != 2 || repo.tabs[0].Sample() || repo.tabs[1].Sample() {
		t.Fatalf("after keep: %+v", repo.tabs)
	}
	// Nothing left to settle: both are no-ops.
	if err := uc.Discard(ctx); err != nil || len(repo.tabs) != 2 {
		t.Fatalf("discard after keep: %v %d", err, len(repo.tabs))
	}
}
