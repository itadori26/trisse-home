package parcels

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/parcels/adapter/controller"
	"trisse-home/internal/modules/parcels/adapter/gateway"
	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/modules/parcels/usecase"
	"trisse-home/internal/shared/kernel"
)

// refreshInterval keeps well below the free DHL quota of 250 calls a day.
const refreshInterval = time.Hour

// Module wires the parcel rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.Parcels
	handlers *controller.Handlers
}

// mailCheck is how often the mail discovery job looks whether the
// configured interval has passed (#166).
const mailCheck = 5 * time.Minute

// New builds the parcels module; cipher encrypts the API key and the mail
// password, dial opens the read-only IMAP connection (#166).
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, dial gateway.DialFunc, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewParcels(usecase.Deps{
		Repo:      gateway.NewRepository(db, cipher),
		Providers: map[domain.Carrier]usecase.Provider{domain.CarrierDHL: gateway.NewDHL(client, gateway.DefaultDHLURL)},
		Clock:     clock, IDs: ids,
		Mail:      gateway.NewIMAP(dial),
		MailStore: gateway.NewMailStore(db, cipher, clock.Now),
	})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "parcels" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"parcels"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/parcels/settings", h.Settings)
	r.Protected("PUT /api/parcels/settings", h.SaveSettings)
	r.Protected("GET /api/parcels/mail", h.MailSettings)
	r.Protected("PUT /api/parcels/mail", h.SaveMailSettings)
	r.Protected("GET /api/parcels", h.List)
	r.Protected("POST /api/parcels", h.Add)
	r.Protected("DELETE /api/parcels/{id}", h.Remove)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{
		{Name: "parcels.refresh", Interval: refreshInterval, Run: m.uc.RefreshAll},
		{Name: "parcels.mail", Interval: mailCheck, Run: m.uc.DiscoverParcels},
	}
}
