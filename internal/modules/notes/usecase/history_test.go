package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/notes/domain"
)

type memRevisions struct{ revs []domain.Revision }

func (m *memRevisions) ListRevisions(_ context.Context, noteID string) ([]domain.Revision, error) {
	var out []domain.Revision
	// Newest first; ties keep the most recent insertion first like the
	// SQLite gateway (created_at DESC, rowid DESC).
	for i := len(m.revs) - 1; i >= 0; i-- {
		if m.revs[i].NoteID == noteID {
			out = append(out, m.revs[i])
		}
	}
	slices.SortStableFunc(out, func(a, b domain.Revision) int { return b.CreatedAt.Compare(a.CreatedAt) })
	return out, nil
}
func (m *memRevisions) AddRevision(_ context.Context, r domain.Revision) error {
	m.revs = append(m.revs, r)
	return nil
}
func (m *memRevisions) DeleteRevisions(_ context.Context, ids ...string) error {
	m.revs = slices.DeleteFunc(m.revs, func(r domain.Revision) bool { return slices.Contains(ids, r.ID) })
	return nil
}

type fakeIndex struct{ got string }

func (f *fakeIndex) Search(_ context.Context, q string, limit int) ([]domain.SearchHit, error) {
	f.got = q
	return []domain.SearchHit{{ID: "x"}}, nil
}

func TestHistoryBacklinksAndSearch(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)}
	revs, index := &memRevisions{}, &fakeIndex{}
	d := Deps{Repo: &memRepo{m: map[string]domain.Note{}}, Clock: c, IDs: &ids{}, Revisions: revs, Search: index}
	save := NewSaveNote(d)

	trip, _ := save.Execute(ctx, "", NoteInput{Title: "Trip", Body: "v1"})
	if len(revs.revs) != 0 {
		t.Fatal("creating a note keeps no revision")
	}
	// The first edit keeps the original; edits within ten minutes do not.
	c.t = c.t.Add(time.Minute)
	_, _ = save.Execute(ctx, trip.ID, NoteInput{Title: "Trip", Body: "v2"})
	c.t = c.t.Add(time.Minute)
	_, _ = save.Execute(ctx, trip.ID, NoteInput{Title: "Trip", Body: "v3"})
	// Pinning alone changes no content.
	c.t = c.t.Add(time.Hour)
	_, _ = save.Execute(ctx, trip.ID, NoteInput{Title: "Trip", Body: "v3", Pinned: true})
	list, err := NewListRevisions(d).Execute(ctx, trip.ID)
	if err != nil || len(list) != 1 {
		t.Fatalf("revisions %+v %v", list, err)
	}
	// After the gap the next edit keeps the content before it.
	_, _ = save.Execute(ctx, trip.ID, NoteInput{Title: "Trip", Body: "v4\nnew line", Pinned: true})
	list, _ = NewListRevisions(d).Execute(ctx, trip.ID)
	if len(list) != 2 {
		t.Fatalf("revisions %+v", list)
	}

	detail, err := NewGetRevision(d).Execute(ctx, trip.ID, list[1].ID)
	if err != nil || detail.Body != "v1" || len(detail.Diff) != 3 || detail.Diff[0].Op != domain.DiffRemove {
		t.Fatalf("detail %+v %v", detail, err)
	}
	if _, err := NewGetRevision(d).Execute(ctx, trip.ID, "nope"); !errors.Is(err, ErrRevisionNotFound) {
		t.Fatalf("unknown revision: %v", err)
	}

	// Restoring keeps the current content as a revision.
	restored, err := NewRestoreRevision(d).Execute(ctx, trip.ID, list[1].ID)
	if err != nil || restored.Body != "v1" || !restored.Pinned {
		t.Fatalf("restore %+v %v", restored, err)
	}
	list, _ = NewListRevisions(d).Execute(ctx, trip.ID)
	if len(list) != 3 {
		t.Fatalf("after restore %+v", list)
	}
	if d, _ := NewGetRevision(d).Execute(ctx, trip.ID, list[0].ID); d.Body != "v4\nnew line" {
		t.Fatalf("restore must keep the replaced content, got %q", d.Body)
	}

	// Backlinks.
	_, _ = save.Execute(ctx, "", NoteInput{Title: "Packing", Body: "For the [[trip]]"})
	_, _ = save.Execute(ctx, "", NoteInput{Title: "Unrelated", Body: "nothing"})
	back, err := NewListBacklinks(d).Execute(ctx, trip.ID)
	if err != nil || len(back) != 1 || back[0].Title != "Packing" {
		t.Fatalf("backlinks %+v %v", back, err)
	}
	if _, err := NewListBacklinks(d).Execute(ctx, "nope"); !errors.Is(err, domain.ErrNoteNotFound) {
		t.Fatalf("unknown note: %v", err)
	}

	// Search builds a safe query and skips empty input.
	if hits, _ := NewSearchNotes(d).Execute(ctx, `trip "x`); len(hits) != 1 || index.got != `"trip"* "x"*` {
		t.Fatalf("search %q", index.got)
	}
	index.got = ""
	if hits, _ := NewSearchNotes(d).Execute(ctx, "  "); len(hits) != 0 || index.got != "" {
		t.Fatal("empty query must not search")
	}
}
