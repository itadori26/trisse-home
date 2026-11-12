package controller

import (
	"net/http"

	"trisse-home/internal/modules/services/adapter/presenter"
	"trisse-home/internal/modules/services/usecase"
	"trisse-home/internal/shared/httpapi"
)

// DiscoveryHandlers serve the Docker autodiscovery endpoints (#78).
type DiscoveryHandlers struct{ uc usecase.Discovery }

// NewDiscoveryHandlers returns discovery handlers.
func NewDiscoveryHandlers(uc usecase.Discovery) *DiscoveryHandlers { return &DiscoveryHandlers{uc: uc} }

// Get handles GET /api/services/discovery.
func (h *DiscoveryHandlers) Get(w http.ResponseWriter, r *http.Request) {
	h.write(w, r)(h.uc.Get(r.Context()))
}

// Save handles PUT /api/services/discovery.
func (h *DiscoveryHandlers) Save(w http.ResponseWriter, r *http.Request) {
	var dto DiscoveryInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	h.write(w, r)(h.uc.Save(r.Context(), dto.toInput()))
}

// Sync handles POST /api/services/discovery/sync.
func (h *DiscoveryHandlers) Sync(w http.ResponseWriter, r *http.Request) {
	h.write(w, r)(h.uc.Sync(r.Context()))
}

func (h *DiscoveryHandlers) write(w http.ResponseWriter, r *http.Request) func(usecase.DiscoveryView, error) {
	return func(v usecase.DiscoveryView, err error) {
		if err != nil {
			httpapi.WriteDomainError(w, r, err)
			return
		}
		httpapi.WriteJSON(w, http.StatusOK, presenter.Discovery(v))
	}
}
