// Package controller contains the new-releases HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/releases/adapter/presenter"
	"trisse-home/internal/modules/releases/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the new-releases endpoints.
type Handlers struct{ uc usecase.Releases }

// NewHandlers returns handlers.
func NewHandlers(uc usecase.Releases) *Handlers { return &Handlers{uc: uc} }

// Search handles GET /api/releases/search?q=….
func (h *Handlers) Search(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.SearchArtists(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Artists(res))
}

// Followed handles GET /api/releases/artists.
func (h *Handlers) Followed(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Followed(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Artists(res))
}

// Follow handles POST /api/releases/artists.
func (h *Handlers) Follow(w http.ResponseWriter, r *http.Request) {
	var dto FollowDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Follow(r.Context(), dto.ID, dto.Name); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unfollow handles DELETE /api/releases/artists/{id}.
func (h *Handlers) Unfollow(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Unfollow(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /api/releases?weeks=N.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	weeks, _ := strconv.Atoi(r.URL.Query().Get("weeks")) // invalid or missing means the default
	res, err := h.uc.List(r.Context(), weeks)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.List(res))
}

// Cover handles GET /api/releases/covers/{id}, proxied from the Cover Art Archive.
func (h *Handlers) Cover(w http.ResponseWriter, r *http.Request) {
	c, err := h.uc.Cover(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", c.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=604800")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(c.Data)
}
