package controller

import (
	"net/http"

	"trisse-home/internal/modules/system/adapter/presenter"
	"trisse-home/internal/modules/system/usecase"
	"trisse-home/internal/shared/httpapi"
)

// VersionHandlers serve the version endpoints.
type VersionHandlers struct{ uc usecase.Versions }

// NewVersionHandlers returns version handlers.
func NewVersionHandlers(uc usecase.Versions) *VersionHandlers { return &VersionHandlers{uc: uc} }

// Version handles GET /api/system/version.
func (h *VersionHandlers) Version(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Info(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Version(res))
}

// SetUpdateCheck handles PUT /api/system/update-check.
func (h *VersionHandlers) SetUpdateCheck(w http.ResponseWriter, r *http.Request) {
	var dto UpdateCheckDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SetCheckEnabled(r.Context(), dto.Enabled)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Version(res))
}
