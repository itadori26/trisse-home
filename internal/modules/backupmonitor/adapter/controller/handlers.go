// Package controller contains the backup monitor HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"io"
	"net/http"
	"time"

	"trisse-home/internal/modules/backupmonitor/adapter/presenter"
	"trisse-home/internal/modules/backupmonitor/domain"
	"trisse-home/internal/modules/backupmonitor/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves the backup monitor endpoints.
type Handlers struct{ uc usecase.Monitor }

// NewHandlers returns backup monitor handlers.
func NewHandlers(uc usecase.Monitor) *Handlers { return &Handlers{uc: uc} }

func (d SourceInputDTO) toInput() usecase.SourceInput {
	return usecase.SourceInput{Name: d.Name, Schedule: domain.Schedule{Kind: d.Kind, Cron: d.Cron,
		MaxAge: time.Duration(d.MaxAgeHours * float64(time.Hour)), Grace: time.Duration(d.GraceMinutes * float64(time.Minute))}}
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// List handles GET /api/backup-status/sources.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Sources(res))
}

// Create handles POST /api/backup-status/sources.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto SourceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Create(r.Context(), dto.toInput())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Created(res, baseURL(r)))
}

// Update handles PUT /api/backup-status/sources/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto SourceInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	res, err := h.uc.Update(r.Context(), r.PathValue("id"), dto.toInput())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Source(res))
}

// Delete handles DELETE /api/backup-status/sources/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RotateToken handles POST /api/backup-status/sources/{id}/token.
func (h *Handlers) RotateToken(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.RotateToken(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.WriteJSON(w, http.StatusOK, presenter.Created(res, baseURL(r)))
}

// Ping handles POST /api/backup-status/ping/{token}[?status=fail]. It is
// public because backup scripts have no session; the unguessable token
// authenticates the call. The optional body is stored as message.
func (h *Handlers) Ping(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, domain.MaxMessageLength))
	event := usecase.PingSuccess
	switch r.PathValue("event") {
	case "start":
		event = usecase.PingStart
	case "fail":
		event = usecase.PingFail
	case "":
		if r.URL.Query().Get("status") == "fail" { // older backup scripts
			event = usecase.PingFail
		}
	default:
		httpapi.WriteDomainError(w, r, domain.ErrSourceNotFound)
		return
	}
	if err := h.uc.Ping(r.Context(), r.PathValue("token"), event, string(body)); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
