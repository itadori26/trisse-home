package weather

import (
	"context"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/weather/adapter/controller"
	"trisse-home/internal/modules/weather/adapter/gateway"
	"trisse-home/internal/modules/weather/usecase"
	"trisse-home/internal/shared/kernel"
)

// refreshInterval keeps cached forecasts fresh; Open-Meteo updates hourly.
const refreshInterval = 15 * time.Minute

// Module wires the weather rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Weather
	handlers *controller.Handlers
	alerts   *controller.AlertHandlers
}

// New builds the weather module. client is used for Open-Meteo requests.
func New(client *http.Client, clock kernel.Clock) *Module {
	provider := gateway.NewOpenMeteo(client, gateway.DefaultForecastURL, gateway.DefaultGeocodingURL, "en")
	uc := usecase.NewWeather(provider, gateway.NewMemoryCache(), clock)
	dwd := gateway.NewDWD(client, gateway.DefaultDWDWFS, gateway.DefaultDWDWMS)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), alerts: controller.NewAlertHandlers(usecase.NewAlerts(dwd, dwd, clock))}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "weather" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"weather", "radar"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/weather", m.handlers.Forecast)
	r.Protected("GET /api/weather/places", m.handlers.Places)
	r.Protected("GET /api/weather/warnings", m.alerts.Warnings)
	r.Protected("GET /api/weather/radar", m.alerts.Radar)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "weather.refresh", Interval: refreshInterval, Run: func(ctx context.Context) error {
		return m.uc.RefreshAll(ctx)
	}}}
}
