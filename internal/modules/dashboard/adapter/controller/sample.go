package controller

import (
	"net/http"

	"waylight/internal/modules/dashboard/usecase"
	"waylight/internal/shared/httpapi"
)

// SampleHandlers settle the example tabs of the setup once the template
// choice on first start is made (#670).
type SampleHandlers struct{ uc usecase.SampleTabs }

// NewSampleHandlers returns the handlers.
func NewSampleHandlers(uc usecase.SampleTabs) *SampleHandlers { return &SampleHandlers{uc: uc} }

// Discard handles DELETE /api/dashboard/sample.
func (h *SampleHandlers) Discard(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Discard(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Keep handles POST /api/dashboard/sample/keep.
func (h *SampleHandlers) Keep(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.Keep(r.Context()); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
