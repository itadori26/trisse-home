// Package controller contains the media server HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/media/adapter/presenter"
	"trisse-home/internal/modules/media/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the media server endpoints.
type Handlers struct{ uc usecase.Media }

// NewHandlers returns media handlers.
func NewHandlers(uc usecase.Media) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/media/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/media/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Overview handles GET /api/media/overview.
func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Overview(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(res))
}

// Poster handles GET /api/media/posters/{id}: the cover is proxied, so the
// browser never talks to the media server or sees its token.
func (h *Handlers) Poster(w http.ResponseWriter, r *http.Request) {
	p, err := h.uc.Poster(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", p.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(p.Data)
}
