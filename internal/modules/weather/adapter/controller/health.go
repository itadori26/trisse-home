package controller

import (
	"net/http"
	"strconv"
	"strings"

	"trisse-home/internal/modules/weather/adapter/presenter"
	"trisse-home/internal/modules/weather/domain"
	"trisse-home/internal/modules/weather/usecase"
	"trisse-home/internal/shared/httpapi"
)

// HealthHandlers serves the pollen and UV forecasts (#56).
type HealthHandlers struct{ uc usecase.Health }

// NewHealthHandlers returns the handlers.
func NewHealthHandlers(uc usecase.Health) *HealthHandlers { return &HealthHandlers{uc: uc} }

// Options handles GET /api/weather/health/options.
func (h *HealthHandlers) Options(w http.ResponseWriter, r *http.Request) {
	o, err := h.uc.Options(r.Context())
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.HealthOptions(o))
}

// Report handles GET /api/weather/health?region=32&city=Hannover&types=birch,grasses.
func (h *HealthHandlers) Report(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	region, _ := strconv.Atoi(q.Get("region"))
	var types []domain.PollenType
	for _, s := range strings.Split(q.Get("types"), ",") {
		if t, ok := domain.ParsePollenType(s); ok {
			types = append(types, t)
		}
	}
	report, err := h.uc.Report(r.Context(), region, strings.TrimSpace(q.Get("city")), types)
	if err != nil {
		httpapi.WriteDomainError(w, r, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, presenter.Health(report))
}
