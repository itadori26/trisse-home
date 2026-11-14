package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"trisse-home/internal/modules/notes/domain"
)

type memRepo struct{ m map[string]domain.Note }

func (r *memRepo) List(context.Context) ([]domain.Note, error) {
	var out []domain.Note
	for _, n := range r.m {
		out = append(out, n)
	}
	return out, nil
}
func (r *memRepo) Get(_ context.Context, id string) (domain.Note, error) {
	n, ok := r.m[id]
	if !ok {
		return n, domain.ErrNoteNotFound
	}
	return n, nil
}
func (r *memRepo) Save(_ context.Context, n domain.Note) error { r.m[n.ID()] = n; return nil }
func (r *memRepo) Delete(_ context.Context, id string) error   { delete(r.m, id); return nil }

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return fmt.Sprint(i.n) }

func TestNotes(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Unix(100, 0)}
	d := Deps{Repo: &memRepo{m: map[string]domain.Note{}}, Clock: c, IDs: &ids{}}
	save := NewSaveNote(d)

	a, err := save.Execute(ctx, "", NoteInput{Body: "# Ideas", Pinned: true})
	if err != nil || a.Title != "Ideas" {
		t.Fatalf("create: %+v %v", a, err)
	}
	if _, err := save.Execute(ctx, "", NoteInput{}); err == nil {
		t.Fatal("expected validation error")
	}
	_, _ = save.Execute(ctx, "", NoteInput{Title: "Other"})
	c.t = time.Unix(200, 0)
	upd, err := save.Execute(ctx, a.ID, NoteInput{Title: "Ideas v2", Body: "x", Pinned: true})
	if err != nil || !upd.UpdatedAt.Equal(time.Unix(200, 0)) || !upd.CreatedAt.Equal(time.Unix(100, 0)) {
		t.Fatalf("update timestamps wrong: %+v", upd)
	}
	pinned, _ := NewListNotes(d).Execute(ctx, true)
	if len(pinned) != 1 {
		t.Fatalf("pinned filter: %d", len(pinned))
	}
	if _, err := save.Execute(ctx, "missing", NoteInput{Title: "x"}); !errors.Is(err, domain.ErrNoteNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := NewDeleteNote(d).Execute(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := NewGetNote(d).Execute(ctx, a.ID); !errors.Is(err, domain.ErrNoteNotFound) {
		t.Fatalf("expected not found after delete, got %v", err)
	}
}

// TestOfflineReplay: creates with a client ID are idempotent and updates
// carry the version they were made from (#80).
func TestOfflineReplay(t *testing.T) {
	ctx := context.Background()
	d := Deps{Repo: &memRepo{m: map[string]domain.Note{}}, Clock: &clk{t: time.Unix(100, 0)}, IDs: &ids{}}
	save := NewSaveNote(d)
	first, err := save.Execute(ctx, "", NoteInput{Body: "Idee", ID: "client-note-1"})
	if err != nil || first.ID != "client-note-1" || first.Version != 1 {
		t.Fatalf("%+v %v", first, err)
	}
	if again, _ := save.Execute(ctx, "", NoteInput{Body: "Idee", ID: "client-note-1"}); again.ID != first.ID || len(d.Repo.(*memRepo).m) != 1 {
		t.Fatalf("replay created a duplicate: %+v", again)
	}
	if _, err := save.Execute(ctx, "", NoteInput{Body: "x", ID: "no"}); err == nil {
		t.Fatal("short client id must be refused")
	}
	if v2, err := save.Execute(ctx, first.ID, NoteInput{Body: "Idee 2", Version: 1}); err != nil || v2.Version != 2 {
		t.Fatalf("%+v %v", v2, err)
	}
	if _, err := save.Execute(ctx, first.ID, NoteInput{Body: "offline", Version: 1}); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("conflict: %v", err)
	}
}
