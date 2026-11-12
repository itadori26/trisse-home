package docker

import (
	"context"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/docker/adapter/controller"
	"trisse-home/internal/modules/docker/adapter/gateway"
	"trisse-home/internal/modules/docker/domain"
	"trisse-home/internal/modules/docker/usecase"
	"trisse-home/internal/shared/kernel"
)

const refreshInterval = 20 * time.Second

// Config configures the docker module.
type Config struct {
	// ProxyURL points to the read-only docker-socket-proxy.
	ProxyURL string
	// Actions enables restart and logs for the listed containers; restarts go
	// through ActionsProxyURL, a second proxy that only allows restarts.
	Actions           bool
	ActionsContainers []string
	ActionsProxyURL   string
}

// Module wires the docker rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Containers
	handlers *controller.Handlers
	actions  *controller.ActionHandlers
}

// New builds the docker module.
func New(client *http.Client, cfg Config, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	read := gateway.NewAPI(client, cfg.ProxyURL)
	uc := usecase.NewContainers(read, sink, clock)
	policy := domain.ActionPolicy{Enabled: cfg.Actions, Allowed: cfg.ActionsContainers}
	actions := usecase.NewActions(policy, gateway.NewAPI(client, cfg.ActionsProxyURL), read, uc.Refresh)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), actions: controller.NewActionHandlers(actions)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "docker" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"docker"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/docker/containers", m.handlers.List)
	r.Protected("GET /api/docker/actions", m.actions.Capabilities)
	r.Protected("POST /api/docker/containers/{name}/restart", m.actions.Restart)
	r.Protected("GET /api/docker/containers/{name}/logs", m.actions.Logs)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "docker.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}

// Metrics implements kernel.MetricSource (#109): container counts and the
// state per container.
func (m *Module) Metrics(ctx context.Context) []kernel.MetricSample {
	res, err := m.uc.List(ctx)
	if err != nil {
		return nil
	}
	const help = "Number of containers by state."
	out := []kernel.MetricSample{
		kernel.Gauge("trisse_containers", help, float64(res.Running), kernel.Label{Name: "state", Value: "running"}),
		kernel.Gauge("trisse_containers", help, float64(res.Stopped), kernel.Label{Name: "state", Value: "stopped"}),
		kernel.Gauge("trisse_containers", help, float64(res.Unhealthy), kernel.Label{Name: "state", Value: "unhealthy"}),
	}
	for _, c := range res.Containers {
		out = append(out, kernel.Gauge("trisse_container_running", "Whether the container is running (1) or not (0).",
			kernel.Bool(c.State == "running"), kernel.Label{Name: "container", Value: c.Name}))
	}
	return out
}
