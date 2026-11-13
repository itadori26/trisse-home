package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/briefing/domain"
	"trisse-home/internal/shared/kernel"
)

// Input ports.

// BuildBriefing composes today's briefing from the source modules (#52).
type BuildBriefing interface {
	Execute(ctx context.Context) (BriefingView, error)
}

// SendBriefing pushes the briefing when the schedule is due (job, #70);
// it reports whether something was sent.
type SendBriefing interface {
	Execute(ctx context.Context) (bool, error)
	// SendNow pushes the briefing immediately (test button).
	SendNow(ctx context.Context) error
}

// ManageSettings reads and changes layout, push schedule and location.
type ManageSettings interface {
	Get(ctx context.Context) (SettingsView, error)
	Save(ctx context.Context, in SettingsInput) (SettingsView, error)
}

// Read ports – owned by this module, implemented in the composition root
// on top of the source modules' input ports (ADR 0016). A nil reader means
// the source module is disabled; its section stays empty.

// TodoFact is an open to-do with a due date.
type TodoFact struct {
	Title   string
	Due     time.Time // date
	Overdue bool
}

// DueTodosReader lists open to-dos due on or before a day.
type DueTodosReader interface {
	DueTodos(ctx context.Context, until time.Time) ([]TodoFact, error)
}

// EventFact is an upcoming appointment.
type EventFact struct {
	Title, Location string
	Start           time.Time
	AllDay          bool
}

// UpcomingEventsReader lists appointments starting before until.
type UpcomingEventsReader interface {
	UpcomingEvents(ctx context.Context, until time.Time) ([]EventFact, error)
}

// WarningFact is an active weather warning; Level 1 (minor) to 4 (extreme).
type WarningFact struct {
	Headline string
	Level    int
	Expires  time.Time
}

// WeatherWarningReader lists active warnings for a location.
type WeatherWarningReader interface {
	WeatherWarnings(ctx context.Context, lat, lon float64) ([]WarningFact, error)
}

// CollectionFact is a waste collection.
type CollectionFact struct {
	Type string
	Day  time.Time // date
}

// CollectionReader lists collections before until.
type CollectionReader interface {
	Collections(ctx context.Context, until time.Time) ([]CollectionFact, error)
}

// ServiceFact is a service that is not online.
type ServiceFact struct {
	Name  string
	Error string
}

// ServiceStatusReader lists services that are down right now.
type ServiceStatusReader interface {
	DownServices(ctx context.Context) ([]ServiceFact, error)
}

// IncidentFact is a recorded failure or warning.
type IncidentFact struct {
	At       time.Time
	Message  string
	Critical bool
}

// IncidentReader lists failures and warnings since a time (timeline, #74).
type IncidentReader interface {
	Incidents(ctx context.Context, since time.Time) ([]IncidentFact, error)
}

// Other output ports.

// Settings is the stored configuration.
type Settings struct {
	Layout   domain.Layout
	Push     domain.PushSchedule
	Lat, Lon float64
	// LastSent is when the briefing was last pushed.
	LastSent time.Time
}

// SettingsRepository stores the settings; Load returns ok=false when none
// were saved yet.
type SettingsRepository interface {
	Load(ctx context.Context) (Settings, bool, error)
	Save(ctx context.Context, s Settings) error
}

// Publisher delivers a pushed briefing (adapter to notifications).
type Publisher interface {
	Publish(ctx context.Context, title, body string, urgent bool)
}

// Sources bundles the optional read ports.
type Sources struct {
	Todos      DueTodosReader
	Events     UpcomingEventsReader
	Warnings   WeatherWarningReader
	Collection CollectionReader
	Services   ServiceStatusReader
	Incidents  IncidentReader
}

// Deps bundles the output ports. DefaultLat/Lon is the weather location
// until one is saved; Location is the user's time zone (nil = time.Local).
type Deps struct {
	Repo                   SettingsRepository
	Clock                  kernel.Clock
	Sources                Sources
	Publisher              Publisher
	Location               *time.Location
	DefaultLat, DefaultLon float64
}

func (d Deps) loc() *time.Location {
	if d.Location == nil {
		return time.Local
	}
	return d.Location
}

func (d Deps) settings(ctx context.Context) (Settings, error) {
	s, ok, err := d.Repo.Load(ctx)
	if err != nil {
		return Settings{}, fmt.Errorf("load settings: %w", err)
	}
	if !ok {
		return Settings{Layout: domain.DefaultLayout(), Push: domain.DefaultPushSchedule(), Lat: d.DefaultLat, Lon: d.DefaultLon}, nil
	}
	s.Layout = domain.NormaliseLayout(s.Layout)
	return s, nil
}

// Views.

// BriefingView is the composed briefing plus the sections whose source
// could not be read.
type BriefingView struct {
	Briefing    domain.Briefing
	Unavailable []domain.SectionKind
	At          time.Time
}

// SettingsView is the configuration at the boundary.
type SettingsView struct {
	Layout    domain.Layout
	PushOn    bool
	PushAt    string
	PushDays  []int
	SkipEmpty bool
	Lat, Lon  float64
	// Available lists the sections whose source module is enabled.
	Available []domain.SectionKind
}

// SettingsInput changes the configuration.
type SettingsInput struct {
	Layout    []domain.LayoutEntry
	PushOn    bool
	PushAt    string
	PushDays  []int
	SkipEmpty bool
	Lat, Lon  float64
}

type briefing struct{ d Deps }

// NewBuildBriefing returns the BuildBriefing use case.
func NewBuildBriefing(d Deps) BuildBriefing { return &briefing{d: d} }

// Execute implements BuildBriefing. Sources are read in parallel; a
// failing source hides its section instead of failing the briefing.
func (uc *briefing) Execute(ctx context.Context) (BriefingView, error) {
	s, err := uc.d.settings(ctx)
	if err != nil {
		return BriefingView{}, err
	}
	return uc.build(ctx, s), nil
}

func (uc *briefing) build(ctx context.Context, s Settings) BriefingView {
	now := uc.d.Clock.Now().In(uc.d.loc())
	readers := uc.readers(s, now)
	items := map[domain.SectionKind][]domain.Item{}
	var unavailable []domain.SectionKind
	var mu sync.Mutex
	var wg sync.WaitGroup
	for kind, read := range readers {
		wg.Go(func() {
			list, err := read(ctx)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				unavailable = append(unavailable, kind)
				return
			}
			items[kind] = list
		})
	}
	wg.Wait()
	// Report unavailable sections in layout order.
	var ordered []domain.SectionKind
	for _, e := range s.Layout {
		for _, k := range unavailable {
			if k == e.Kind && e.Enabled {
				ordered = append(ordered, k)
			}
		}
	}
	return BriefingView{Briefing: domain.Compose(s.Layout, items), Unavailable: ordered, At: now}
}

type reader func(ctx context.Context) ([]domain.Item, error)

// readers maps every available section to a function that reads and
// converts its source.
func (uc *briefing) readers(s Settings, now time.Time) map[domain.SectionKind]reader {
	src := uc.d.Sources
	out := map[domain.SectionKind]reader{}
	tomorrow := domain.EndOfTomorrow(now)
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	if src.Todos != nil {
		out[domain.SectionTodos] = func(ctx context.Context) ([]domain.Item, error) {
			facts, err := src.Todos.DueTodos(ctx, today)
			return mapFacts(facts, err, func(f TodoFact) domain.Item {
				it := domain.Item{Text: f.Title, At: f.Due}
				if f.Overdue {
					it.Detail, it.Severity = "overdue", domain.SeverityWarning
				}
				return it
			})
		}
	}
	if src.Events != nil {
		out[domain.SectionEvents] = func(ctx context.Context) ([]domain.Item, error) {
			facts, err := src.Events.UpcomingEvents(ctx, tomorrow)
			return mapFacts(facts, err, func(f EventFact) domain.Item {
				return domain.Item{Text: f.Title, Detail: f.Location, At: f.Start}
			})
		}
	}
	if src.Warnings != nil {
		out[domain.SectionWarnings] = func(ctx context.Context) ([]domain.Item, error) {
			facts, err := src.Warnings.WeatherWarnings(ctx, s.Lat, s.Lon)
			return mapFacts(facts, err, func(f WarningFact) domain.Item {
				it := domain.Item{Text: f.Headline, At: f.Expires, Severity: domain.SeverityWarning}
				if f.Level >= 3 {
					it.Severity = domain.SeverityCritical
				}
				return it
			})
		}
	}
	if src.Collection != nil {
		out[domain.SectionWaste] = func(ctx context.Context) ([]domain.Item, error) {
			facts, err := src.Collection.Collections(ctx, tomorrow)
			return mapFacts(facts, err, func(f CollectionFact) domain.Item {
				detail := "tomorrow"
				if !f.Day.After(today) {
					detail = "today"
				}
				return domain.Item{Text: f.Type, Detail: detail, At: f.Day}
			})
		}
	}
	if src.Services != nil {
		out[domain.SectionServices] = func(ctx context.Context) ([]domain.Item, error) {
			facts, err := src.Services.DownServices(ctx)
			return mapFacts(facts, err, func(f ServiceFact) domain.Item {
				return domain.Item{Text: f.Name, Detail: f.Error, Severity: domain.SeverityCritical}
			})
		}
	}
	if src.Incidents != nil {
		out[domain.SectionIncidents] = func(ctx context.Context) ([]domain.Item, error) {
			facts, err := src.Incidents.Incidents(ctx, domain.IncidentsSince(now))
			return mapFacts(facts, err, func(f IncidentFact) domain.Item {
				it := domain.Item{Text: f.Message, At: f.At, Severity: domain.SeverityWarning}
				if f.Critical {
					it.Severity = domain.SeverityCritical
				}
				return it
			})
		}
	}
	return out
}

func mapFacts[F any](facts []F, err error, conv func(F) domain.Item) ([]domain.Item, error) {
	if err != nil {
		return nil, err
	}
	out := make([]domain.Item, 0, len(facts))
	for _, f := range facts {
		out = append(out, conv(f))
	}
	return out, nil
}

// available lists the sections whose source is wired.
func (d Deps) available() []domain.SectionKind {
	src := d.Sources
	has := map[domain.SectionKind]bool{
		domain.SectionTodos: src.Todos != nil, domain.SectionEvents: src.Events != nil,
		domain.SectionWarnings: src.Warnings != nil, domain.SectionWaste: src.Collection != nil,
		domain.SectionServices: src.Services != nil, domain.SectionIncidents: src.Incidents != nil,
	}
	var out []domain.SectionKind
	for _, k := range domain.DefaultOrder {
		if has[k] {
			out = append(out, k)
		}
	}
	return out
}

type settings struct{ d Deps }

// NewManageSettings returns the ManageSettings use case.
func NewManageSettings(d Deps) ManageSettings { return &settings{d: d} }

func (uc *settings) view(s Settings) SettingsView {
	return SettingsView{Layout: s.Layout, PushOn: s.Push.Enabled, PushAt: s.Push.At(), PushDays: s.Push.DayNumbers(),
		SkipEmpty: s.Push.SkipEmpty, Lat: s.Lat, Lon: s.Lon, Available: uc.d.available()}
}

// Get implements ManageSettings.
func (uc *settings) Get(ctx context.Context) (SettingsView, error) {
	s, err := uc.d.settings(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	return uc.view(s), nil
}

// Save implements ManageSettings.
func (uc *settings) Save(ctx context.Context, in SettingsInput) (SettingsView, error) {
	current, err := uc.d.settings(ctx)
	if err != nil {
		return SettingsView{}, err
	}
	push, err := domain.NewPushSchedule(in.PushOn, in.PushAt, in.PushDays, in.SkipEmpty)
	if err != nil {
		return SettingsView{}, err
	}
	if in.Lat < -90 || in.Lat > 90 || in.Lon < -180 || in.Lon > 180 {
		return SettingsView{}, kernel.NewValidationError("location", "Latitude must be between -90 and 90, longitude between -180 and 180.")
	}
	s := Settings{Layout: domain.NormaliseLayout(in.Layout), Push: push, Lat: in.Lat, Lon: in.Lon, LastSent: current.LastSent}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	return uc.view(s), nil
}

type send struct {
	d     Deps
	build *briefing
}

// NewSendBriefing returns the SendBriefing use case; it reuses the same
// composition as the widget, only the delivery channel differs (#70).
func NewSendBriefing(d Deps) SendBriefing { return &send{d: d, build: &briefing{d: d}} }

// Execute implements SendBriefing.
func (uc *send) Execute(ctx context.Context) (bool, error) {
	if uc.d.Publisher == nil {
		return false, nil
	}
	s, err := uc.d.settings(ctx)
	if err != nil {
		return false, err
	}
	now := uc.d.Clock.Now().In(uc.d.loc())
	if !s.Push.Due(now, s.LastSent) {
		return false, nil
	}
	// Remember the attempt first, so a failing source never leads to a
	// push every minute.
	s.LastSent = now
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return false, fmt.Errorf("save settings: %w", err)
	}
	v := uc.build.build(ctx, s)
	if v.Briefing.Empty() && s.Push.SkipEmpty {
		return false, nil
	}
	uc.publish(ctx, v)
	return true, nil
}

// SendNow implements SendBriefing.
func (uc *send) SendNow(ctx context.Context) error {
	if uc.d.Publisher == nil {
		return kernel.NewPublicError(kernel.ErrUnavailable, "Notifications are not enabled.")
	}
	s, err := uc.d.settings(ctx)
	if err != nil {
		return err
	}
	uc.publish(ctx, uc.build.build(ctx, s))
	return nil
}

func (uc *send) publish(ctx context.Context, v BriefingView) {
	uc.d.Publisher.Publish(ctx, "Briefing for "+v.At.Format("Mon, 2 Jan"), v.Briefing.Text(uc.d.loc()),
		v.Briefing.MostSevere() == domain.SeverityCritical)
}
