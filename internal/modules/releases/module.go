package releases

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/releases/adapter/controller"
	"trisse-home/internal/modules/releases/adapter/gateway"
	"trisse-home/internal/modules/releases/usecase"
	"trisse-home/internal/shared/kernel"
)

// refreshInterval: new releases appear at most daily; with one request per
// second per artist a refresh of 50 artists takes about a minute.
const refreshInterval = 6 * time.Hour

// Module wires the new-releases rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.Releases
	handlers *controller.Handlers
}

// New builds the module (#63).
func New(db *sql.DB, client *http.Client, clock kernel.Clock) *Module {
	uc := usecase.NewReleases(gateway.NewArtistRepository(db),
		gateway.NewMusicBrainz(client, gateway.DefaultMusicBrainzURL, gateway.DefaultCoverArtURL), clock)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "releases" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"releases"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/releases", m.handlers.List)
	r.Protected("GET /api/releases/search", m.handlers.Search)
	r.Protected("GET /api/releases/artists", m.handlers.Followed)
	r.Protected("POST /api/releases/artists", m.handlers.Follow)
	r.Protected("DELETE /api/releases/artists/{id}", m.handlers.Unfollow)
	r.Protected("GET /api/releases/covers/{id}", m.handlers.Cover)
}

// Jobs implements modules.Module; nothing is requested while no artist is followed.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "releases.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}
