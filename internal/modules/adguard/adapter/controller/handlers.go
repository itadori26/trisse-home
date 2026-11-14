package controller

import (
	"net/http"

	"trisse-home/internal/modules/adguard/adapter/presenter"
	"trisse-home/internal/modules/adguard/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the DNS filter endpoints. Every endpoint takes the
// optional query parameter provider=adguard|pihole (default adguard).
type Handlers struct{ uc usecase.DNSFilter }

// NewHandlers returns DNS filter handlers.
func NewHandlers(uc usecase.DNSFilter) *Handlers { return &Handlers{uc: uc} }

func provider(r *http.Request) string { return r.URL.Query().Get("provider") }

// Settings handles GET /api/adguard/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Settings(r.Context(), provider(r))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// SaveSettings handles PUT /api/adguard/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.SaveSettings(r.Context(), usecase.SettingsInput{Provider: provider(r), URL: dto.URL, Username: dto.Username, Password: dto.Password})
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Settings(res))
}

// Stats handles GET /api/adguard/stats.
func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Stats(r.Context(), provider(r))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Stats(res))
}

// Pause handles POST /api/adguard/pause (#177).
func (h *Handlers) Pause(w http.ResponseWriter, r *http.Request) {
	var dto PauseDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Pause(r.Context(), provider(r), dto.Minutes)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Stats(res))
}
