package domain

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestWikiLinks(t *testing.T) {
	body := "See [[Shopping list]] and [[ Trip 2027 | the trip]].\n" +
		"Again [[shopping LIST]].\n" +
		"`[[not a link]]`\n```\n[[also not]]\n```\n[[]] [[broken\nlink]]"
	if got := WikiLinks(body); !reflect.DeepEqual(got, []string{"Shopping list", "Trip 2027"}) {
		t.Fatalf("got %q", got)
	}
	n, _ := NewNote("n1", "A", body, false, time.Now())
	if !n.LinksTo("shopping list") || n.LinksTo("also not") {
		t.Fatal("LinksTo")
	}
}

func TestRevisionRules(t *testing.T) {
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	if ShouldRecordRevision(false, time.Time{}, now) {
		t.Fatal("unchanged content")
	}
	if !ShouldRecordRevision(true, time.Time{}, now) {
		t.Fatal("first revision")
	}
	if ShouldRecordRevision(true, now.Add(-9*time.Minute), now) {
		t.Fatal("burst of edits")
	}
	if !ShouldRecordRevision(true, now.Add(-RevisionGap), now) {
		t.Fatal("after the gap")
	}
	var revs []Revision
	for i := range MaxRevisions + 3 {
		revs = append(revs, Revision{ID: fmt.Sprint(i)})
	}
	if got := RevisionsToPrune(revs); !reflect.DeepEqual(got, []string{"50", "51", "52"}) {
		t.Fatalf("prune %v", got)
	}
	if RevisionsToPrune(revs[:10]) != nil {
		t.Fatal("nothing to prune")
	}
}

func TestDiffLines(t *testing.T) {
	got := DiffLines("a\nb\nc\nd", "a\nc\nd\ne")
	want := []DiffLine{{DiffSame, "a"}, {DiffRemove, "b"}, {DiffSame, "c"}, {DiffSame, "d"}, {DiffAdd, "e"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
	if d := DiffLines("", "x"); len(d) != 2 || d[0].Op != DiffRemove || d[1].Op != DiffAdd {
		t.Fatalf("empty to text %v", d)
	}
}

func TestSearchQuery(t *testing.T) {
	for in, want := range map[string]string{
		`büro schlüssel`:        `"büro"* "schlüssel"*`,
		`"quoted" OR x*`:        `"quoted"* "OR"* "x"*`,
		`  `:                    ``,
		`a b c d e f g h i j k`: `"a"* "b"* "c"* "d"* "e"* "f"* "g"* "h"* "i"* "j"*`,
	} {
		if got := SearchQuery(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
