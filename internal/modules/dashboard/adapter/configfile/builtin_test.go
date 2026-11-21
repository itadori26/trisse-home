package configfile

import (
	"slices"
	"testing"
)

// webCells are the grid cells of the web sizes (#595), as the web client
// sizes a tile; a sized widget of a template must have exactly these.
var webCells = map[string][2]int{"small": {3, 2}, "medium": {4, 4}, "large": {6, 6}, "strip": {6, 1}}

// TestBuiltin checks the shipped templates (#589): each file parses with
// known fields only, has an ID and a title, places its widgets on the grid
// without overlap and gives every widget a phone slot. Each template shows
// the zoned layout (#619, #670): zones on, every widget in a zone, a status
// widget in the System zone, and sized widgets with the cells of their size.
func TestBuiltin(t *testing.T) {
	list, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, b := range list {
		ids = append(ids, b.ID)
		if b.ID == "" || b.Title == "" || b.Tab.Name == "" || len(b.Tab.Widgets) == 0 {
			t.Errorf("%s: incomplete %+v", b.ID, b)
		}
		if !b.Tab.Zones {
			t.Errorf("%s: not zoned", b.ID)
		}
		system := false
		for i, w := range b.Tab.Widgets {
			r := w.Rect
			if w.Zone == "" {
				t.Errorf("%s: %s has no zone", b.ID, w.Kind)
			}
			system = system || w.Zone == "system"
			if cells, ok := webCells[w.Size]; ok && (r.W != cells[0] || r.H != cells[1]) {
				t.Errorf("%s: %s is %s but %dx%d", b.ID, w.Kind, w.Size, r.W, r.H)
			}
			if r.X < 0 || r.W < 1 || r.X+r.W > 12 || r.H < 1 {
				t.Errorf("%s: %s outside the grid: %+v", b.ID, w.Kind, r)
			}
			if w.Phone == nil {
				t.Errorf("%s: %s has no phone slot", b.ID, w.Kind)
			}
			for _, o := range b.Tab.Widgets[i+1:] {
				q := o.Rect
				if r.X < q.X+q.W && q.X < r.X+r.W && r.Y < q.Y+q.H && q.Y < r.Y+r.H {
					t.Errorf("%s: %s overlaps %s", b.ID, w.Kind, o.Kind)
				}
			}
		}
		if !system {
			t.Errorf("%s: nothing in the System zone", b.ID)
		}
	}
	if !slices.Equal(ids, []string{"family", "homelab", "wall", "work"}) {
		t.Fatalf("templates: %v", ids)
	}
}

func TestMergeConfig(t *testing.T) {
	got, err := MergeConfig(`{"days":4,"lat":1}`, map[string]any{"lat": 53.1, "label": "Home"})
	if err != nil || got != `{"days":4,"label":"Home","lat":53.1}` {
		t.Fatalf("%s %v", got, err)
	}
	if _, err := MergeConfig(`not json`, nil); err == nil {
		t.Fatal("broken settings")
	}
}
