package integrations

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"maps"
	"slices"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/integrations/adapter/configfile"
	"waylight/internal/modules/integrations/adapter/controller"
	"waylight/internal/modules/integrations/adapter/gateway"
	"waylight/internal/modules/integrations/domain"
	"waylight/internal/modules/integrations/usecase"
	"waylight/internal/shared/kernel"
)

// Widgets lists widgets with the connection their settings name, so a
// connection can say which widgets use it.
type Widgets func(ctx context.Context) ([]usecase.Widget, error)

// Module wires the Integrations rings and implements modules.Module and
// modules.Configurable. It is not Portable: a data export never carries
// credentials, and connections without them are of no use.
type Module struct {
	modules.Base
	repo     usecase.Repository
	specs    domain.Specs
	handlers *controller.Handlers
	config   usecase.Config
	copier   usecase.Copier
	sources  []modules.ConnectionSource
	managed  func() bool
	log      *slog.Logger
}

// Deps are the dependencies of the module.
type Deps struct {
	DB *sql.DB
	// Cipher encrypts credentials with SECRET_KEY.
	Cipher gateway.Cipher
	// Modules are the other enabled modules: their connection specs and
	// the settings to copy (modules.Connectable, modules.ConnectionSource).
	Modules []modules.Module
	Widgets Widgets
	// Managed reports whether config as code owns the integrations
	// section; copies wait while it does, so the file stays the only
	// source of the connection list. Nil means never.
	Managed func() bool
	Log     *slog.Logger
	Clock   kernel.Clock
	IDs     kernel.IDGenerator
}

// New builds the integrations module (#610).
func New(d Deps) *Module {
	repo := gateway.NewRepository(d.DB, d.Cipher)
	svc := services(modules.ConnectionSpecsOf(d.Modules))
	uc := usecase.New(usecase.Deps{Repo: repo, Widgets: d.Widgets, Services: svc, Clock: d.Clock, IDs: d.IDs})
	specs := make([]domain.Spec, 0, len(svc))
	for _, s := range svc {
		specs = append(specs, s.Spec)
	}
	m := &Module{repo: repo, specs: domain.NewSpecs(specs), handlers: controller.NewHandlers(uc), config: usecase.NewConfig(repo, d.Clock, d.IDs),
		copier: usecase.NewCopier(repo, svc, d.Clock, d.IDs), managed: d.Managed, log: d.Log}
	for _, o := range d.Modules {
		if s, ok := o.(modules.ConnectionSource); ok {
			m.sources = append(m.sources, s)
		}
	}
	return m
}

// copyInterval is how often settings saved in a module after the first
// run are looked for; a run without anything new only reads the settings.
const copyInterval = time.Hour

// Jobs implements modules.Module: copy the module settings into
// connections, at start and then hourly (#610).
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "integrations-copy-settings", Interval: copyInterval, Run: m.copySettings}}
}

// copySettings reads the settings of every module and copies the new ones.
// A module whose settings cannot be read (e.g. after a SECRET_KEY change)
// is skipped until the next run. Only sources and reasons are logged,
// never a credential.
func (m *Module) copySettings(ctx context.Context) error {
	if m.managed != nil && m.managed() {
		return nil
	}
	var saved []usecase.Saved
	for _, src := range m.sources {
		list, err := src.SavedConnections(ctx)
		if err != nil {
			m.log.WarnContext(ctx, "integrations: module settings not readable, not copied", "module", moduleName(src), "err", err)
			continue
		}
		for _, s := range list {
			saved = append(saved, usecase.Saved{Source: s.Source, Service: s.Service, Name: s.Name, URL: s.URL, Credential: s.Credential})
		}
	}
	res, err := m.copier.Copy(ctx, saved)
	if err != nil {
		return err
	}
	if len(res.Copied) > 0 {
		m.log.InfoContext(ctx, "integrations: copied module settings into connections; the module settings stay in use", "sources", res.Copied)
	}
	for _, src := range slices.Sorted(maps.Keys(res.Skipped)) {
		m.log.WarnContext(ctx, "integrations: module settings not copied, will try again", "source", src, "reason", res.Skipped[src])
	}
	return nil
}

func moduleName(src modules.ConnectionSource) string {
	if n, ok := src.(interface{ Name() string }); ok {
		return n.Name()
	}
	return "unknown"
}

// services converts the module specs into the ports of the use cases, so
// the inner rings do not depend on the module contract.
func services(specs []modules.ConnectionSpec) []usecase.Service {
	out := make([]usecase.Service, 0, len(specs))
	for _, s := range specs {
		spec := domain.Spec{Service: s.Service, Title: s.Title, FixedURL: s.FixedURL}
		for _, f := range s.Fields {
			spec.Fields = append(spec.Fields, domain.Field{Name: f.Name, Secret: f.Secret, Optional: f.Optional})
		}
		out = append(out, usecase.Service{Spec: spec, Check: s.Check})
	}
	return out
}

// Name implements modules.Module.
func (m *Module) Name() string { return "integrations" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind says connections belong to the instance and its admins.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/integrations/services", m.handlers.Services)
	r.Protected("GET /api/integrations/connections", m.handlers.List)
	r.Protected("POST /api/integrations/connections/{id}/check", m.handlers.Check)
	r.Protected("POST /api/integrations/connections", m.handlers.Add)
	r.Protected("PUT /api/integrations/connections/{id}", m.handlers.Update)
	r.Protected("PUT /api/integrations/connections/{id}/credential", m.handlers.ReplaceCredential)
	r.Protected("DELETE /api/integrations/connections/{id}", m.handlers.Remove)
}

// ConfigSection implements modules.Configurable.
func (m *Module) ConfigSection() string { return "integrations" }

// ExportConfig implements modules.Configurable.
func (m *Module) ExportConfig(ctx context.Context) (any, error) {
	list, err := m.config.Export(ctx)
	if err != nil {
		return nil, err
	}
	return configfile.FromList(list), nil
}

// ConfigRoutes implements modules.Configurable. Replacing a credential is
// not listed: the file names connections, the credential is entered here.
func (m *Module) ConfigRoutes() []string {
	return []string{
		"POST /api/integrations/connections",
		"PUT /api/integrations/connections/{id}",
		"DELETE /api/integrations/connections/{id}",
	}
}

// NewConfig implements modules.Configurable.
func (m *Module) NewConfig() any { return &configfile.Section{} }

// ApplyConfig implements modules.Configurable.
func (m *Module) ApplyConfig(ctx context.Context, section any) error {
	s, ok := section.(*configfile.Section)
	if !ok {
		return modules.ErrConfigType
	}
	return m.config.Apply(ctx, s.List())
}

// ResolveConnection implements modules.ConnectionResolver: the address and
// credential of a connection of the service. A connection of another
// service, or one whose credential is incomplete (e.g. no longer readable
// after a SECRET_KEY change), answers modules.ErrNoConnection, so the
// module falls back to its own settings.
func (m *Module) ResolveConnection(ctx context.Context, service, id string) (modules.ResolvedConnection, error) {
	c, err := m.repo.Get(ctx, id)
	if err != nil || c.Service != service {
		return modules.ResolvedConnection{}, modules.ErrNoConnection
	}
	if spec, ok := m.specs[service]; !ok || spec.CheckCredential(c.Credential, true) != nil {
		return modules.ResolvedConnection{}, modules.ErrNoConnection
	}
	return modules.ResolvedConnection{Name: c.Name, URL: c.URL, Credential: maps.Clone(c.Credential)}, nil
}
