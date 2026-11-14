package controller

import (
	"net/http"

	"trisse-home/internal/modules/notes/adapter/presenter"
	"trisse-home/internal/modules/notes/usecase"
	"trisse-home/internal/shared/httpapi"
)

// UseCases bundles the notes input ports.
type UseCases struct {
	List   usecase.ListNotes
	Get    usecase.GetNote
	Save   usecase.SaveNote
	Delete usecase.DeleteNote
	// #73
	Search    usecase.SearchNotes
	Backlinks usecase.ListBacklinks
	Revisions usecase.ListRevisions
	Revision  usecase.GetRevision
	Restore   usecase.RestoreRevision
}

// Handlers serves the notes endpoints.
type Handlers struct{ uc UseCases }

// NewHandlers returns notes handlers.
func NewHandlers(uc UseCases) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/notes?pinned=true.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List.Execute(r.Context(), r.URL.Query().Get("pinned") == "true")
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Notes(res))
}

// Get handles GET /api/notes/{id}.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Get.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Note(res))
}

// Create handles POST /api/notes.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	h.save(w, r, "", http.StatusCreated)
}

// Update handles PUT /api/notes/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	h.save(w, r, r.PathValue("id"), http.StatusOK)
}

func (h *Handlers) save(w http.ResponseWriter, r *http.Request, id string, status int) {
	var dto NoteInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Save.Execute(r.Context(), id, usecase.NoteInput{Title: dto.Title, Body: dto.Body, Pinned: dto.Pinned, ID: dto.ID, Version: dto.Version})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, status, presenter.Note(res))
}

// Search handles GET /api/notes/search?q=….
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Search.Execute(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SearchHits(res))
}

// Backlinks handles GET /api/notes/{id}/backlinks.
func (h *Handlers) Backlinks(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Backlinks.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.NoteRefs(res))
}

// Revisions handles GET /api/notes/{id}/revisions.
func (h *Handlers) Revisions(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Revisions.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Revisions(res))
}

// Revision handles GET /api/notes/{id}/revisions/{rev}.
func (h *Handlers) Revision(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Revision.Execute(r.Context(), r.PathValue("id"), r.PathValue("rev"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.RevisionDetail(res))
}

// Restore handles POST /api/notes/{id}/revisions/{rev}/restore.
func (h *Handlers) Restore(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Restore.Execute(r.Context(), r.PathValue("id"), r.PathValue("rev"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Note(res))
}

// Delete handles DELETE /api/notes/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete.Execute(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
