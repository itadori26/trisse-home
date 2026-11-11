package docker

import (
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/docker/adapter/controller"
	"trisse-home/internal/modules/docker/adapter/gateway"
	"trisse-home/internal/modules/docker/usecase"
	"trisse-home/internal/shared/kernel"
)

const refreshInterval = 20 * time.Second

// Module wires the docker rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Containers
	handlers *controller.Handlers
}

// New builds the docker module. proxyURL points to docker-socket-proxy.
func New(client *http.Client, proxyURL string, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	uc := usecase.NewContainers(gateway.NewAPI(client, proxyURL), sink, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "docker" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"docker"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/docker/containers", m.handlers.List)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "docker.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
