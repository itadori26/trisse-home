// Package controller contains the timeline HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"strconv"
	"time"

	"trisse-home/internal/modules/timeline/adapter/presenter"
	"trisse-home/internal/modules/timeline/domain"
	"trisse-home/internal/modules/timeline/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the timeline endpoints.
type Handlers struct{ uc usecase.Timeline }

// NewHandlers returns timeline handlers.
func NewHandlers(uc usecase.Timeline) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/timeline?source=&severity=&before=<RFC 3339>&limit=.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := usecase.Filter{Source: q.Get("source"), Severity: domain.Severity(q.Get("severity"))}
	f.Limit, _ = strconv.Atoi(q.Get("limit")) // invalid or missing means the default
	if b := q.Get("before"); b != "" {
		t, err := time.Parse(time.RFC3339, b)
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "before must be an RFC 3339 timestamp.")
			return
		}
		f.Before = t
	}
	list, err := h.uc.List(r.Context(), f)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Events(list))
}

// Settings handles GET /api/timeline/settings.
func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	d, err := h.uc.RetentionDays(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SettingsDTO{RetentionDays: d})
}

// SaveSettings handles PUT /api/timeline/settings.
func (h *Handlers) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var dto SettingsDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	if err := h.uc.SetRetentionDays(r.Context(), dto.RetentionDays); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.SettingsDTO(dto))
}
