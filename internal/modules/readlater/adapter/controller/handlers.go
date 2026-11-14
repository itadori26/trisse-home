// Package controller contains the read-later HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"io"
	"net/http"
	"strconv"

	"trisse-home/internal/modules/readlater/adapter/presenter"
	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/modules/readlater/usecase"
	"trisse-home/internal/shared/httpapi"
)

// SaveDTO is the body of POST /api/readlater.
type SaveDTO struct {
	URL string `json:"url"`
}

// StatusDTO is the body of PUT /api/readlater/{id}/status.
type StatusDTO struct {
	Status string `json:"status"`
}

// Handlers serves /api/readlater.
type Handlers struct{ uc usecase.ReadLater }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.ReadLater) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/readlater?status=unread|read|archived (none = not archived).
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	status, _ := domain.ParseStatus(r.URL.Query().Get("status"))
	as, err := h.uc.List(r.Context(), status)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Articles(as))
}

// Save handles POST /api/readlater.
func (h *Handlers) Save(w http.ResponseWriter, r *http.Request) {
	var dto SaveDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	a, err := h.uc.Save(r.Context(), dto.URL)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	a.Content = ""
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Article(a))
}

// Get handles GET /api/readlater/{id} with the reader copy.
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	a, err := h.uc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Article(a))
}

// SetStatus handles PUT /api/readlater/{id}/status.
func (h *Handlers) SetStatus(w http.ResponseWriter, r *http.Request) {
	var dto StatusDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.SetStatus(r.Context(), r.PathValue("id"), domain.Status(dto.Status)); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/readlater/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Image handles GET /api/readlater/{id}/images/{n}: a stored raster image.
func (h *Handlers) Image(w http.ResponseWriter, r *http.Request) {
	n, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		httpapi.WriteDomainError(w, r, domain.ErrNotFound)
		return
	}
	img, err := h.uc.Image(r.Context(), r.PathValue("id"), n)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", img.ContentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	_, _ = w.Write(img.Bytes)
}

// Settings handles GET /api/readlater/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	s, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SettingsDTO{ArchiveDays: s.ArchiveDays})
}

// SaveSettings handles PUT /api/readlater/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	s, err := h.uc.SaveSettings(r.Context(), dto.ArchiveDays)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SettingsDTO{ArchiveDays: s.ArchiveDays})
}

// ImportPocket handles POST /api/readlater/import/pocket with the CSV file as body.
func (h *Handlers) ImportPocket(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, usecase.MaxImportBytes))
	if err != nil {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The export file must be at most 5 MB.")
		return
	}
	rep, err := h.uc.ImportPocket(r.Context(), data)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ImportReportDTO(rep))
}
