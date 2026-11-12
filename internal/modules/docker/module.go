package docker

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
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

// updateCheckInterval: the job runs hourly; each image is asked at its
// registry at most every 6 hours (domain.UpdateCheckInterval).
const updateCheckInterval = time.Hour

// Config configures the docker module.
type Config struct {
	// ProxyURL points to the read-only docker-socket-proxy.
	ProxyURL string
	// Actions enables restart and logs for the listed containers; restarts go
	// through ActionsProxyURL, a second proxy that only allows restarts.
	Actions           bool
	ActionsContainers []string
	ActionsProxyURL   string
	// DB and Cipher store the image update hints (#107); Registry is the
	// HTTP client for registry requests (only when the user enabled them).
	DB       *sql.DB
	Cipher   gateway.Cipher
	Registry *http.Client
}

// Module wires the docker rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Containers
	updates  usecase.ImageUpdates
	handlers *controller.Handlers
	actions  *controller.ActionHandlers
	upd      *controller.UpdateHandlers
}

// New builds the docker module.
func New(client *http.Client, cfg Config, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	read := gateway.NewAPI(client, cfg.ProxyURL)
	uc := usecase.NewContainers(read, sink, clock)
	policy := domain.ActionPolicy{Enabled: cfg.Actions, Allowed: cfg.ActionsContainers}
	actions := usecase.NewActions(policy, gateway.NewAPI(client, cfg.ActionsProxyURL), read, uc.Refresh)
	store := gateway.NewUpdateStore(cfg.DB, cfg.Cipher)
	updates := usecase.NewImageUpdates(usecase.UpdateDeps{Images: read, Registry: gateway.NewRegistry(cfg.Registry),
		Cache: store, Settings: store, Sink: sink, Clock: clock})
	return &Module{uc: uc, updates: updates, handlers: controller.NewHandlers(uc), actions: controller.NewActionHandlers(actions),
		upd: controller.NewUpdateHandlers(updates)}
}

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

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
	r.Protected("GET /api/docker/updates", m.upd.List)
	r.Protected("GET /api/docker/updates/settings", m.upd.Settings)
	r.Protected("PUT /api/docker/updates/settings", m.upd.SaveSettings)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{
		{Name: "docker.refresh", Interval: refreshInterval, Run: m.uc.Refresh},
		// Does nothing until the update check is enabled.
		{Name: "docker.image-updates", Interval: updateCheckInterval, Run: m.updates.Check},
	}
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

// ErrContainersUnavailable is returned by Containers while the Docker API
// is not reachable.
var ErrContainersUnavailable = errors.New("docker api not reachable")

// Containers is a read model for other modules (wired in cmd/server): the
// cached container list with labels. It fails while the Docker API is not
// reachable, so readers never mistake a stale or empty list for the truth.
func (m *Module) Containers(ctx context.Context) ([]usecase.ContainerView, error) {
	res, err := m.uc.List(ctx)
	if err != nil {
		return nil, err
	}
	if res.Error != "" {
		return nil, ErrContainersUnavailable
	}
	return res.Containers, nil
}
