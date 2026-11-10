// Package controller contains the bookmarks HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/bookmarks/adapter/presenter"
	"trisse-home/internal/modules/bookmarks/usecase"
	"trisse-home/internal/shared/httpapi"
)

// UseCases bundles the bookmarks input ports.
type UseCases struct {
	List         usecase.ListBookmarks
	Create       usecase.CreateBookmark
	Update       usecase.UpdateBookmark
	Delete       usecase.DeleteBookmark
	CreateFolder usecase.CreateFolder
	RenameFolder usecase.RenameFolder
	DeleteFolder usecase.DeleteFolder
}

// Handlers serves the bookmarks endpoints.
type Handlers struct{ uc UseCases }

// NewHandlers returns bookmarks handlers.
func NewHandlers(uc UseCases) *Handlers { return &Handlers{uc: uc} }

func (d BookmarkInputDTO) toInput() usecase.Input {
	return usecase.Input{FolderID: d.FolderID, Title: d.Title, URL: d.URL, Description: d.Description, Tags: d.Tags}
}

// List handles GET /api/bookmarks?folder=…&tag=…&q=….
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	res, err := h.uc.List.Execute(r.Context(), usecase.Filter{FolderID: q.Get("folder"), Tag: q.Get("tag"), Query: q.Get("q")})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.List(res))
}

// Create handles POST /api/bookmarks.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto BookmarkInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Create.Execute(r.Context(), dto.toInput())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Bookmark(res))
}

// Update handles PUT /api/bookmarks/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto BookmarkInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Update.Execute(r.Context(), r.PathValue("id"), dto.toInput())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Bookmark(res))
}

// Delete handles DELETE /api/bookmarks/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete.Execute(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateFolder handles POST /api/bookmarks/folders.
func (h *Handlers) CreateFolder(w http.ResponseWriter, r *http.Request) {
	var dto FolderInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.CreateFolder.Execute(r.Context(), dto.Name)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Folder(res))
}

// RenameFolder handles PATCH /api/bookmarks/folders/{id}.
func (h *Handlers) RenameFolder(w http.ResponseWriter, r *http.Request) {
	var dto FolderInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.RenameFolder.Execute(r.Context(), r.PathValue("id"), dto.Name)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Folder(res))
}

// DeleteFolder handles DELETE /api/bookmarks/folders/{id}.
func (h *Handlers) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteFolder.Execute(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
