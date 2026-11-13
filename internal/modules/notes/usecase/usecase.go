package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/notes/domain"
	"trisse-home/internal/shared/kernel"
)

// NoteInput carries the editable fields of a note.
type NoteInput struct {
	Title  string
	Body   string
	Pinned bool
}

// NoteView is a note at the use case boundary.
type NoteView struct {
	ID, Title, Body      string
	Pinned               bool
	CreatedAt, UpdatedAt time.Time
}

// ListNotes returns all notes, pinned first, then most recently edited (input port).
type ListNotes interface {
	Execute(ctx context.Context, pinnedOnly bool) ([]NoteView, error)
}

// GetNote returns one note (input port).
type GetNote interface {
	Execute(ctx context.Context, id string) (NoteView, error)
}

// SaveNote creates a note (id == "") or updates an existing one (input port).
type SaveNote interface {
	Execute(ctx context.Context, id string, in NoteInput) (NoteView, error)
}

// DeleteNote removes a note (input port).
type DeleteNote interface {
	Execute(ctx context.Context, id string) error
}

// SearchNotes runs the full-text search (input port, #73).
type SearchNotes interface {
	Execute(ctx context.Context, query string) ([]domain.SearchHit, error)
}

// ListBacklinks returns the notes that link to a note with [[Title]] (input port, #73).
type ListBacklinks interface {
	Execute(ctx context.Context, id string) ([]NoteRef, error)
}

// ListRevisions returns the earlier versions of a note, newest first (input port, #73).
type ListRevisions interface {
	Execute(ctx context.Context, id string) ([]RevisionView, error)
}

// GetRevision returns an earlier version and its diff to the current note (input port, #73).
type GetRevision interface {
	Execute(ctx context.Context, id, revisionID string) (RevisionDetail, error)
}

// RestoreRevision makes an earlier version current again; the current
// content becomes a revision itself (input port, #73).
type RestoreRevision interface {
	Execute(ctx context.Context, id, revisionID string) (NoteView, error)
}

// NoteRef names a note.
type NoteRef struct{ ID, Title string }

// RevisionView describes a revision without its body.
type RevisionView struct {
	ID, Title string
	CreatedAt time.Time
	Size      int
}

// RevisionDetail is a revision with its body and the diff from the
// revision to the current note.
type RevisionDetail struct {
	RevisionView
	Body string
	Diff []domain.DiffLine
}

// SearchIndex is the full-text index of the notes (output port, #73).
type SearchIndex interface {
	Search(ctx context.Context, ftsQuery string, limit int) ([]domain.SearchHit, error)
}

// RevisionRepository stores earlier versions of notes (output port, #73).
type RevisionRepository interface {
	// ListRevisions returns the revisions of a note, newest first.
	ListRevisions(ctx context.Context, noteID string) ([]domain.Revision, error)
	AddRevision(ctx context.Context, r domain.Revision) error
	DeleteRevisions(ctx context.Context, ids ...string) error
}

// Repository persists notes (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Note, error)
	Get(ctx context.Context, id string) (domain.Note, error)
	Save(ctx context.Context, n domain.Note) error
	Delete(ctx context.Context, id string) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo  Repository
	Clock kernel.Clock
	IDs   kernel.IDGenerator
	// Search and Revisions are optional (nil = feature off).
	Search    SearchIndex
	Revisions RevisionRepository
}

func toView(n domain.Note) NoteView {
	return NoteView{ID: n.ID(), Title: n.Title(), Body: n.Body(), Pinned: n.Pinned(), CreatedAt: n.CreatedAt(), UpdatedAt: n.UpdatedAt()}
}

type listNotes struct{ d Deps }

// NewListNotes returns the ListNotes use case.
func NewListNotes(d Deps) ListNotes { return &listNotes{d: d} }

// Execute implements ListNotes; the repository already orders the notes.
func (uc *listNotes) Execute(ctx context.Context, pinnedOnly bool) ([]NoteView, error) {
	notes, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	out := make([]NoteView, 0, len(notes))
	for _, n := range notes {
		if !pinnedOnly || n.Pinned() {
			out = append(out, toView(n))
		}
	}
	return out, nil
}

type getNote struct{ d Deps }

// NewGetNote returns the GetNote use case.
func NewGetNote(d Deps) GetNote { return &getNote{d: d} }

// Execute implements GetNote.
func (uc *getNote) Execute(ctx context.Context, id string) (NoteView, error) {
	n, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return NoteView{}, err
	}
	return toView(n), nil
}

type saveNote struct{ d Deps }

// NewSaveNote returns the SaveNote use case.
func NewSaveNote(d Deps) SaveNote { return &saveNote{d: d} }

// Execute implements SaveNote.
func (uc *saveNote) Execute(ctx context.Context, id string, in NoteInput) (NoteView, error) {
	now := uc.d.Clock.Now()
	var n domain.Note
	if id == "" {
		created, err := domain.NewNote(uc.d.IDs.NewID(), in.Title, in.Body, in.Pinned, now)
		if err != nil {
			return NoteView{}, err
		}
		n = created
	} else {
		existing, err := uc.d.Repo.Get(ctx, id)
		if err != nil {
			return NoteView{}, err
		}
		before := existing
		if err := existing.Edit(in.Title, in.Body, in.Pinned, now); err != nil {
			return NoteView{}, err
		}
		if err := uc.d.keepRevision(ctx, before, existing.ContentChanged(before.Title(), before.Body()), false); err != nil {
			return NoteView{}, err
		}
		n = existing
	}
	if err := uc.d.Repo.Save(ctx, n); err != nil {
		return NoteView{}, fmt.Errorf("save note: %w", err)
	}
	return toView(n), nil
}

type deleteNote struct{ d Deps }

// NewDeleteNote returns the DeleteNote use case.
func NewDeleteNote(d Deps) DeleteNote { return &deleteNote{d: d} }

// Execute implements DeleteNote.
func (uc *deleteNote) Execute(ctx context.Context, id string) error {
	if _, err := uc.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	if err := uc.d.Repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}
