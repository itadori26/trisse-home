package summary

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/summary/adapter/controller"
	"trisse-home/internal/modules/summary/adapter/gateway"
	"trisse-home/internal/modules/summary/usecase"
	"trisse-home/internal/shared/kernel"
)

// Config are the dependencies of the module.
type Config struct {
	DB *sql.DB
	// Model talks to the local Ollama (integration guard, long timeout).
	Model *http.Client
	// Pages loads feed articles (internal address guard).
	Pages *http.Client
	Clock kernel.Clock
}

// Module wires the summary rings (#75) and implements modules.Module. It
// has no widget: notes and the feed reader show a "Summarise" button once
// an Ollama server is set up.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the module; it makes no request until Ollama is configured.
func New(c Config) *Module {
	uc := usecase.NewSummary(usecase.Deps{Repo: gateway.NewRepository(c.DB), Model: gateway.NewOllama(c.Model),
		Fetcher: gateway.NewArticle(c.Pages), Clock: c.Clock})
	return &Module{Transferable: modules.Transferable{DB: c.DB, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "summary" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/summary/settings", m.handlers.Settings)
	r.Protected("PUT /api/summary/settings", m.handlers.SaveSettings)
	r.Protected("DELETE /api/summary/settings", m.handlers.Disable)
	r.Protected("POST /api/summary/models", m.handlers.Models)
	r.Protected("POST /api/summary", m.handlers.Summarize)
	r.Protected("POST /api/summary/ask", m.handlers.Ask)
}
