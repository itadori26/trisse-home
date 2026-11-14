package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/notes/domain"
	"trisse-home/internal/shared/kernel"
)

// maxSearchHits limits the full-text search results.
const maxSearchHits = 20

// ErrRevisionNotFound is returned for unknown revision IDs.
var ErrRevisionNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The version does not exist.")

// keepRevision stores the previous content of a note as a revision when
// the rules ask for it (or always when force is set) and prunes old ones.
func (d Deps) keepRevision(ctx context.Context, before domain.Note, changed, force bool) error {
	if d.Revisions == nil || !changed {
		return nil
	}
	revs, err := d.Revisions.ListRevisions(ctx, before.ID())
	if err != nil {
		return fmt.Errorf("list revisions: %w", err)
	}
	now := d.Clock.Now()
	var newest time.Time // zero: the first edit always keeps the original
	if len(revs) > 0 {
		newest = revs[0].CreatedAt
	}
	if !force && !domain.ShouldRecordRevision(true, newest, now) {
		return nil
	}
	rev := domain.Revision{ID: d.IDs.NewID(), NoteID: before.ID(), Title: before.Title(), Body: before.Body(), CreatedAt: now}
	if err := d.Revisions.AddRevision(ctx, rev); err != nil {
		return fmt.Errorf("add revision: %w", err)
	}
	if prune := domain.RevisionsToPrune(append([]domain.Revision{rev}, revs...)); len(prune) > 0 {
		if err := d.Revisions.DeleteRevisions(ctx, prune...); err != nil {
			return fmt.Errorf("prune revisions: %w", err)
		}
	}
	return nil
}

func revisionView(r domain.Revision) RevisionView {
	return RevisionView{ID: r.ID, Title: r.Title, CreatedAt: r.CreatedAt, Size: len(r.Body)}
}

func (d Deps) findRevision(ctx context.Context, noteID, revID string) (domain.Revision, error) {
	if d.Revisions == nil {
		return domain.Revision{}, ErrRevisionNotFound
	}
	revs, err := d.Revisions.ListRevisions(ctx, noteID)
	if err != nil {
		return domain.Revision{}, fmt.Errorf("list revisions: %w", err)
	}
	for _, r := range revs {
		if r.ID == revID {
			return r, nil
		}
	}
	return domain.Revision{}, ErrRevisionNotFound
}

type searchNotes struct{ d Deps }

// NewSearchNotes returns the SearchNotes use case.
func NewSearchNotes(d Deps) SearchNotes { return &searchNotes{d: d} }

// Execute implements SearchNotes.
func (uc *searchNotes) Execute(ctx context.Context, query string) ([]domain.SearchHit, error) {
	q := domain.SearchQuery(query)
	if q == "" || uc.d.Search == nil {
		return []domain.SearchHit{}, nil
	}
	hits, err := uc.d.Search.Search(ctx, q, maxSearchHits)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}
	return hits, nil
}

type listBacklinks struct{ d Deps }

// NewListBacklinks returns the ListBacklinks use case.
func NewListBacklinks(d Deps) ListBacklinks { return &listBacklinks{d: d} }

// Execute implements ListBacklinks.
func (uc *listBacklinks) Execute(ctx context.Context, id string) ([]NoteRef, error) {
	target, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	notes, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	out := []NoteRef{}
	for _, n := range notes {
		if n.ID() != id && n.LinksTo(target.Title()) {
			out = append(out, NoteRef{ID: n.ID(), Title: n.Title()})
		}
	}
	return out, nil
}

type listRevisions struct{ d Deps }

// NewListRevisions returns the ListRevisions use case.
func NewListRevisions(d Deps) ListRevisions { return &listRevisions{d: d} }

// Execute implements ListRevisions.
func (uc *listRevisions) Execute(ctx context.Context, id string) ([]RevisionView, error) {
	if _, err := uc.d.Repo.Get(ctx, id); err != nil {
		return nil, err
	}
	out := []RevisionView{}
	if uc.d.Revisions == nil {
		return out, nil
	}
	revs, err := uc.d.Revisions.ListRevisions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list revisions: %w", err)
	}
	for _, r := range revs {
		out = append(out, revisionView(r))
	}
	return out, nil
}

type getRevision struct{ d Deps }

// NewGetRevision returns the GetRevision use case.
func NewGetRevision(d Deps) GetRevision { return &getRevision{d: d} }

// Execute implements GetRevision.
func (uc *getRevision) Execute(ctx context.Context, id, revID string) (RevisionDetail, error) {
	n, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return RevisionDetail{}, err
	}
	r, err := uc.d.findRevision(ctx, id, revID)
	if err != nil {
		return RevisionDetail{}, err
	}
	return RevisionDetail{RevisionView: revisionView(r), Body: r.Body, Diff: domain.DiffLines(r.Body, n.Body())}, nil
}

type restoreRevision struct{ d Deps }

// NewRestoreRevision returns the RestoreRevision use case.
func NewRestoreRevision(d Deps) RestoreRevision { return &restoreRevision{d: d} }

// Execute implements RestoreRevision.
func (uc *restoreRevision) Execute(ctx context.Context, id, revID string) (NoteView, error) {
	n, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return NoteView{}, err
	}
	r, err := uc.d.findRevision(ctx, id, revID)
	if err != nil {
		return NoteView{}, err
	}
	before := n
	if err := n.Edit(r.Title, r.Body, n.Pinned(), uc.d.Clock.Now()); err != nil {
		return NoteView{}, err
	}
	// The current content is kept, so a restore can be undone.
	if err := uc.d.keepRevision(ctx, before, n.ContentChanged(before.Title(), before.Body()), true); err != nil {
		return NoteView{}, err
	}
	n.Bump()
	if err := uc.d.Repo.Save(ctx, n); err != nil {
		return NoteView{}, fmt.Errorf("save note: %w", err)
	}
	return toView(n), nil
}
