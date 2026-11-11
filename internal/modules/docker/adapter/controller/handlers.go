package controller

import (
	"net/http"

	"trisse-home/internal/modules/docker/adapter/presenter"
	"trisse-home/internal/modules/docker/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves GET /api/docker/containers.
type Handlers struct{ uc usecase.Containers }

// NewHandlers returns docker handlers.
func NewHandlers(uc usecase.Containers) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/docker/containers.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.List(res))
}

// ActionHandlers serves the opt-in container actions.
type ActionHandlers struct{ uc usecase.Actions }

// NewActionHandlers returns action handlers.
func NewActionHandlers(uc usecase.Actions) *ActionHandlers { return &ActionHandlers{uc: uc} }

// Capabilities handles GET /api/docker/actions.
func (h *ActionHandlers) Capabilities(w http.ResponseWriter, _ *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, presenter.Capabilities(h.uc.Capabilities()))
}

// Restart handles POST /api/docker/containers/{name}/restart.
func (h *ActionHandlers) Restart(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Restart(r.Context(), r.PathValue("name")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Logs handles GET /api/docker/containers/{name}/logs.
func (h *ActionHandlers) Logs(w http.ResponseWriter, r *http.Request) {
	lines, err := h.uc.Logs(r.Context(), r.PathValue("name"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if lines == nil {
		lines = []string{}
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusOK, presenter.LogsDTO{Lines: lines})
}
