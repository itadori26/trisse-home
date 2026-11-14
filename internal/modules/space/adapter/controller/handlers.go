// Package controller contains the space HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"

	"trisse-home/internal/modules/space/adapter/presenter"
	"trisse-home/internal/modules/space/usecase"
	"trisse-home/internal/shared/httpapi"
)

// SettingsInputDTO is the body of PUT /api/space/settings; an empty nasaKey
// with keepKey keeps the stored key.
type SettingsInputDTO struct {
	Launches bool   `json:"launches"`
	Picture  bool   `json:"picture"`
	NASAKey  string `json:"nasaKey"`
	KeepKey  bool   `json:"keepKey"`
}

// Handlers serves /api/space.
type Handlers struct{ uc usecase.Space }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Space) *Handlers { return &Handlers{uc: uc} }

// Launches handles GET /api/space/launches?limit=5.
func (h *Handlers) Launches(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		n = 5
	}
	ls, err := h.uc.Launches(r.Context(), n)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Launches(ls))
}

// Picture handles GET /api/space/apod.
func (h *Handlers) Picture(w http.ResponseWriter, r *http.Request) {
	p, err := h.uc.Picture(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Picture(p))
}

// Image handles GET /api/space/apod/image: today's picture, proxied.
func (h *Handlers) Image(w http.ResponseWriter, r *http.Request) {
	img, err := h.uc.PictureImage(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", img.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(img.Bytes)
}

// Settings handles GET /api/space/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s))
}

// SaveSettings handles PUT /api/space/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.uc.SaveSettings(r.Context(), dto.Launches, dto.Picture, dto.NASAKey, dto.KeepKey)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(s))
}
