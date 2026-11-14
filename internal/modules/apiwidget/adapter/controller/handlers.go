// Package controller contains the HTTP handlers of the JSON widget.
//
// Ring: Interface Adapters.
package controller

import (
	"net/http"
	"time"

	"trisse-home/internal/modules/apiwidget/adapter/presenter"
	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/modules/apiwidget/usecase"
	"trisse-home/internal/shared/httpapi"
)

// Handlers serves /api/apiwidget.
type Handlers struct{ uc usecase.APIWidget }

// NewHandlers returns the handlers.
func NewHandlers(uc usecase.APIWidget) *Handlers { return &Handlers{uc: uc} }

func input(d EndpointInputDTO) domain.EndpointInput {
	in := domain.EndpointInput{Name: d.Name, URL: d.URL, HeaderName: d.HeaderName, HeaderValue: d.HeaderValue,
		KeepHeader: d.KeepHeader, Interval: time.Duration(d.IntervalSeconds) * time.Second, AllowPublic: d.AllowPublic}
	for _, x := range d.Extractions {
		in.Extractions = append(in.Extractions, domain.Extraction(x))
	}
	return in
}

// List handles GET /api/apiwidget/endpoints.
func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	es, err := h.uc.Endpoints(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Endpoints(es))
}

// Create handles POST /api/apiwidget/endpoints.
func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var dto EndpointInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	e, err := h.uc.CreateEndpoint(r.Context(), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, presenter.Endpoint(e))
}

// Update handles PUT /api/apiwidget/endpoints/{id}.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	var dto EndpointInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	e, err := h.uc.UpdateEndpoint(r.Context(), r.PathValue("id"), input(dto))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Endpoint(e))
}

// Delete handles DELETE /api/apiwidget/endpoints/{id}.
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.uc.DeleteEndpoint(r.Context(), r.PathValue("id")); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Preview handles POST /api/apiwidget/preview.
func (h *Handlers) Preview(w http.ResponseWriter, r *http.Request) {
	var dto PreviewInputDTO
	if err := httpapi.DecodeJSON(w, r, &dto); err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	p, err := h.uc.Preview(r.Context(), dto.ID, input(dto.Endpoint))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Preview(p))
}

// Values handles GET /api/apiwidget/endpoints/{id}/values.
func (h *Handlers) Values(w http.ResponseWriter, r *http.Request) {
	v, err := h.uc.FetchValues(r.Context(), r.PathValue("id"))
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Values(v))
}
