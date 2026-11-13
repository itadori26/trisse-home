package weather

import (
	"context"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/weather/adapter/controller"
	"trisse-home/internal/modules/weather/adapter/gateway"
	"trisse-home/internal/modules/weather/domain"
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
	health   *controller.HealthHandlers
	warnings usecase.Alerts
}

// New builds the weather module. client is used for Open-Meteo requests.
func New(client *http.Client, clock kernel.Clock) *Module {
	provider := gateway.NewOpenMeteo(client, gateway.DefaultForecastURL, gateway.DefaultGeocodingURL, "en")
	uc := usecase.NewWeather(provider, gateway.NewMemoryCache(), clock)
	dwd := gateway.NewDWD(client, gateway.DefaultDWDWFS, gateway.DefaultDWDWMS)
	alerts := usecase.NewAlerts(dwd, dwd, clock)
	health := usecase.NewHealth(gateway.NewHealth(dwd, gateway.DefaultPollenURL, gateway.DefaultUVURL), clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), alerts: controller.NewAlertHandlers(alerts),
		health: controller.NewHealthHandlers(health), warnings: alerts}
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
	r.Protected("GET /api/weather/health/options", m.health.Options)
	r.Protected("GET /api/weather/health", m.health.Report)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "weather.refresh", Interval: refreshInterval, Run: func(ctx context.Context) error {
		return m.uc.RefreshAll(ctx)
	}}}
}

// Warnings returns the active weather warnings for a location (read
// access for the briefing, ADR 0016).
func (m *Module) Warnings(ctx context.Context, lat, lon float64) ([]domain.Warning, error) {
	return m.warnings.Warnings(ctx, lat, lon)
}
