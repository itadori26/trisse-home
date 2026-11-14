// Package controller contains the study HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"

	"trisse-home/internal/modules/study/adapter/presenter"
	"trisse-home/internal/modules/study/domain"
	"trisse-home/internal/modules/study/usecase"
	"trisse-home/internal/shared/httpapi"
)

// ModuleInputDTO is the body of POST/PUT /api/study/modules.
type ModuleInputDTO struct {
	Name     string   `json:"name"`
	ECTS     float64  `json:"ects"`
	Semester int      `json:"semester"`
	Status   string   `json:"status"`
	Grade    *float64 `json:"grade"`
}

// SettingsInputDTO is the body of PUT /api/study/settings.
type SettingsInputDTO struct {
	TotalECTS float64 `json:"totalEcts"`
}

// Handlers serves /api/study.
type Handlers struct{ uc usecase.Study }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.Study) *Handlers { return &Handlers{uc: uc} }

func input(d ModuleInputDTO) domain.ModuleInput {
	return domain.ModuleInput{Name: d.Name, ECTS: d.ECTS, Semester: d.Semester, Status: domain.Status(d.Status), Grade: d.Grade}
}

// List handles GET /api/study/modules.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	ms, err := h.uc.Modules(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Modules(ms))
}

// Create handles POST /api/study/modules.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto ModuleInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	m, err := h.uc.Create(r.Context(), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Module(m))
}

// Update handles PUT /api/study/modules/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto ModuleInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	m, err := h.uc.Update(r.Context(), r.PathValue("id"), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Module(m))
}

// Delete handles DELETE /api/study/modules/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Progress handles GET /api/study/progress.
func (h *Handlers) Progress(w http.ResponseWriter, r *http.Request) {
	p, ms, err := h.uc.Progress(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Progress(p, ms))
}

// Settings handles GET /api/study/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.TotalECTS(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SettingsDTO{TotalECTS: v})
}

// SaveSettings handles PUT /api/study/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.SetTotalECTS(r.Context(), dto.TotalECTS)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SettingsDTO{TotalECTS: v})
}
