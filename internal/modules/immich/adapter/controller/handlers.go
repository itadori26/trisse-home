// Package controller contains the Immich HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/immich/adapter/presenter"
	"trisse-home/internal/modules/immich/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the Immich endpoints.
type Handlers struct{ uc usecase.Immich }

// NewHandlers returns Immich handlers.
func NewHandlers(uc usecase.Immich) *Handlers { return &Handlers{uc: uc} }

// Settings handles GET /api/immich/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/immich/settings.
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

// Overview handles GET /api/immich/overview.
func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Overview(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Overview(res))
}

// Thumbnail handles GET /api/immich/thumbnails/{id}: memories are proxied,
// so the browser never talks to Immich or sees the API key.
func (h *Handlers) Thumbnail(w http.ResponseWriter, r *http.Request) {
	t, err := h.uc.Thumbnail(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", t.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(t.Data)
}
