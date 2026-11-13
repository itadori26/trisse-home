package gateway

import (
	"context"
	"reflect"
	"testing"
	"time"

	"trisse-home/internal/modules/notes/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func text(segs []domain.Segment) (all, hits string) {
	for _, s := range segs {
		all += s.Text
		if s.Hit {
			hits += "[" + s.Text + "]"
		}
	}
	return all, hits
}

func TestFullTextSearchFollowsEdits(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "notes", Migrations()))
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	a, _ := domain.NewNote("a", "Büro", "Der Schlüssel liegt im Büro unter der Pflanze.", false, now)
	b, _ := domain.NewNote("b", "Einkauf", "Milch, Brot und ein neuer Schlüsselbund", false, now)
	for _, n := range []domain.Note{a, b} {
		if err := repo.Save(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	search := func(q string) []domain.SearchHit {
		t.Helper()
		hits, err := repo.Search(ctx, domain.SearchQuery(q), 10)
		if err != nil {
			t.Fatal(err)
		}
		return hits
	}
	// Diacritics are folded and words are prefix-matched.
	hits := search("schlussel")
	if len(hits) != 2 {
		t.Fatalf("hits %+v", hits)
	}
	// Title matches rank first.
	hits = search("buro")
	if len(hits) != 1 || hits[0].ID != "a" {
		t.Fatalf("hits %+v", hits)
	}
	if all, marked := text(hits[0].Title); all != "Büro" || marked != "[Büro]" {
		t.Fatalf("title %q %q", all, marked)
	}
	if _, marked := text(hits[0].Snippet); marked != "[Büro]" {
		t.Fatalf("snippet %q", marked)
	}

	// Updates and deletes keep the index in sync.
	_ = b.Edit("Einkauf", "Nur noch Milch", false, now)
	if err := repo.Save(ctx, b); err != nil {
		t.Fatal(err)
	}
	if hits := search("schlüsselbund"); len(hits) != 0 {
		t.Fatalf("stale index %+v", hits)
	}
	if err := repo.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if hits := search("büro"); len(hits) != 0 {
		t.Fatalf("deleted note found %+v", hits)
	}
}

func TestSegments(t *testing.T) {
	got := segments("…a \x02b\x03 c\x02d\x03")
	want := []domain.Segment{{Text: "…a "}, {Text: "b", Hit: true}, {Text: " c"}, {Text: "d", Hit: true}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if got := segments(""); len(got) != 0 {
		t.Fatalf("empty %+v", got)
	}
}

func TestRevisionStorage(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "notes", Migrations()))
	now := time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)
	n, _ := domain.NewNote("n", "T", "v1", false, now)
	if err := repo.Save(ctx, n); err != nil {
		t.Fatal(err)
	}
	for i, body := range []string{"v1", "v2", "v3"} {
		r := domain.Revision{ID: body, NoteID: "n", Title: "T", Body: body, CreatedAt: now.Add(time.Duration(i) * time.Hour)}
		if err := repo.AddRevision(ctx, r); err != nil {
			t.Fatal(err)
		}
	}
	revs, err := repo.ListRevisions(ctx, "n")
	if err != nil || len(revs) != 3 || revs[0].ID != "v3" || revs[2].Body != "v1" {
		t.Fatalf("revisions %+v %v", revs, err)
	}
	if err := repo.DeleteRevisions(ctx, "v1", "v2"); err != nil {
		t.Fatal(err)
	}
	if revs, _ := repo.ListRevisions(ctx, "n"); len(revs) != 1 {
		t.Fatalf("after delete %+v", revs)
	}
	if err := repo.Delete(ctx, "n"); err != nil {
		t.Fatal(err)
	}
	if revs, _ := repo.ListRevisions(ctx, "n"); len(revs) != 0 {
		t.Fatal("revisions must go with the note")
	}
}
