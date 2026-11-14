package modules

import (
	"context"
	"io/fs"
	"net/http"
	"slices"
	"time"

	"trisse-home/internal/shared/httpapi"
)

// Router lets a module register its HTTP handlers. Patterns use the Go 1.22+
// ServeMux syntax, e.g. "GET /api/bookmarks/{id}".
type Router interface {
	// Public registers a handler that does not require a session.
	Public(pattern string, h http.HandlerFunc)
	// Protected registers a handler that requires a valid session and, for
	// state-changing methods, a valid CSRF token.
	Protected(pattern string, h http.HandlerFunc)
}

// Job is a recurring background task, e.g. refreshing a cache.
type Job struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

// Module is the contract between a feature module and the platform.
type Module interface {
	// Name is the unique, lowercase identifier used in configuration.
	Name() string
	// Migrations returns the module's embedded SQL migrations or nil.
	Migrations() fs.FS
	// RegisterRoutes adds the module's HTTP handlers.
	RegisterRoutes(r Router)
	// Jobs returns the module's background jobs.
	Jobs() []Job
	// WidgetKinds lists the dashboard widget kinds the module backs.
	WidgetKinds() []string
}

// Seeder is implemented by modules that can fill a fresh installation with
// sample data. SeedSampleData must be idempotent: it only adds data when the
// module has none yet, so a reset and a second setup create no duplicates.
type Seeder interface {
	SeedSampleData(ctx context.Context) error
}

// GuestGated is implemented by modules with read endpoints that are too
// personal for every guest of the module (#168): such a route is readable by
// a guest only when the named widget kind is on one of the guest's tabs.
// Keys are path prefixes ("/api/fritzbox/calls"), values widget kinds.
type GuestGated interface {
	GuestGatedRoutes() map[string]string
}

// Base provides no-op defaults so modules only implement what they need.
type Base struct{}

// Migrations returns nil.
func (Base) Migrations() fs.FS { return nil }

// RegisterRoutes registers nothing.
func (Base) RegisterRoutes(Router) {}

// Jobs returns no jobs.
func (Base) Jobs() []Job { return nil }

// WidgetKinds returns no widget kinds.
func (Base) WidgetKinds() []string { return nil }

// Registry holds the enabled modules in registration order.
type Registry struct {
	modules []Module
}

// Register adds m to the registry.
func (r *Registry) Register(m Module) { r.modules = append(r.modules, m) }

// All returns the registered modules in registration order.
func (r *Registry) All() []Module { return slices.Clone(r.modules) }

// Names returns the names of all registered modules.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.modules))
	for _, m := range r.modules {
		names = append(names, m.Name())
	}
	return names
}

// Seeders returns all registered modules that provide sample data.
func (r *Registry) Seeders() map[string]Seeder {
	out := map[string]Seeder{}
	for _, m := range r.modules {
		if s, ok := m.(Seeder); ok {
			out[m.Name()] = s
		}
	}
	return out
}

// Portables returns all registered modules whose data can be exported,
// keyed by module name.
func (r *Registry) Portables() map[string]Portable {
	out := map[string]Portable{}
	for _, m := range r.modules {
		if p, ok := m.(Portable); ok {
			out[m.Name()] = p
		}
	}
	return out
}

// WidgetKinds returns every widget kind backed by a registered module.
func (r *Registry) WidgetKinds() []string {
	var kinds []string
	for _, m := range r.modules {
		kinds = append(kinds, m.WidgetKinds()...)
	}
	return kinds
}

// Handler returns a handler that lists the enabled modules so the frontend
// can hide features that are switched off.
func (r *Registry) Handler(staticKinds []string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		kinds := append(slices.Clone(staticKinds), r.WidgetKinds()...)
		httpapi.WriteJSON(w, http.StatusOK, InfoResponse{Modules: r.Names(), WidgetKinds: kinds})
	}
}
