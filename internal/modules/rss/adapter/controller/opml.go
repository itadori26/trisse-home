package controller

import (
	"errors"
	"io"
	"net/http"

	"trisse-home/internal/modules/rss/adapter/presenter"
	"trisse-home/internal/modules/rss/domain"
	"trisse-home/internal/shared/httpapi"
)

// ExportOPML handles GET /api/rss/opml: all subscriptions as an OPML 2.0
// download (#97).
func (h *Handlers) ExportOPML(w http.ResponseWriter, r *http.Request) {
	data, err := h.uc.ExportOPML(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/x-opml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="trisse-home-feeds.opml"`)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

// ImportOPML handles POST /api/rss/opml with the OPML file as body (at most 2 MB).
func (h *Handlers) ImportOPML(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxImportBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "too_large", "The OPML file is larger than 2 MB.")
			return
		}
		httpapi.WriteError(w, http.StatusBadRequest, "bad_request", "The file could not be read.")
		return
	}
	report, err := h.uc.ImportOPML(r.Context(), data)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.ImportReport(report))
}
