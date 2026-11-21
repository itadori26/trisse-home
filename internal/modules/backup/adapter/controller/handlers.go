// Package controller contains the backup HTTP handlers.
//
// Ring: Interface Adapters.
package controller

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"waylight/internal/modules/backup/adapter/presenter"
	"waylight/internal/modules/backup/usecase"
	"waylight/internal/shared/httpapi"
)

// Handlers serves the backup endpoints.
type Handlers struct{ uc usecase.Backups }

// NewHandlers returns backup handlers.
func NewHandlers(uc usecase.Backups) *Handlers { return &Handlers{uc: uc} }

// List handles GET /api/backups.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.List(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Files(res))
}

// Create handles POST /api/backups.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.BackupNow(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.File(res))
}

// Status handles GET /api/backups/status.
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Status(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Status(res))
}

// Download handles GET /api/backups/{name}.
func (h *Handlers) Download(w http.ResponseWriter, r *http.Request) {
	rc, f, err := h.uc.Open(r.Context(), r.PathValue("name"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", "application/vnd.sqlite3")
	w.Header().Set("Content-Disposition", `attachment; filename="`+f.Name+`"`) // name is validated, contains no quotes
	w.Header().Set("Content-Length", strconv.FormatInt(f.Size, 10))
	w.Header().Set("Cache-Control", "no-store")
	if _, err := io.Copy(w, rc); err != nil {
		slog.WarnContext(r.Context(), "backup download interrupted", "err", err)
	}
}
