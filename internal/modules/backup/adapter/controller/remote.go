package controller

import (
	"net/http"

	"trisse-home/internal/modules/backup/adapter/presenter"
	"trisse-home/internal/modules/backup/usecase"
	"trisse-home/internal/shared/httpapi"
)

// RemoteHandlers serves the off-site backup endpoints (#46).
type RemoteHandlers struct{ uc usecase.Offsite }

// NewRemoteHandlers returns off-site handlers.
func NewRemoteHandlers(uc usecase.Offsite) *RemoteHandlers { return &RemoteHandlers{uc: uc} }

func (dto RemoteSettingsInputDTO) input() usecase.RemoteSettingsInput {
	return usecase.RemoteSettingsInput{URL: dto.URL, Username: dto.Username, Password: dto.Password, Retention: dto.Retention}
}

// Settings handles GET /api/backups/remote/settings.
func (h *RemoteHandlers) Settings(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.Settings(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.RemoteSettings(v))
}

// SaveSettings handles PUT /api/backups/remote/settings.
func (h *RemoteHandlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto RemoteSettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	v, err := h.uc.SaveSettings(r.Context(), dto.input())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.RemoteSettings(v))
}

// Test handles POST /api/backups/remote/test.
func (h *RemoteHandlers) Test(w http.ResponseWriter, r *http.Request) {
	var dto RemoteSettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.Test(r.Context(), dto.input()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Disable handles DELETE /api/backups/remote/settings.
func (h *RemoteHandlers) Disable(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Disable(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /api/backups/remote.
func (h *RemoteHandlers) List(w http.ResponseWriter, r *http.Request) {
	files, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Files(files))
}
