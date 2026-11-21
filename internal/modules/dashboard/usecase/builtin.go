package usecase

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
)

// BuiltinTemplate is a tab template that ships with Waylight (#589).
type BuiltinTemplate struct {
	ID, Title string
	// Needs lists what the onboarding asks before applying it; "location"
	// is the only value so far.
	Needs []string
	Tab   ConfigTab
}

// Location is a place for the widgets of a template that show weather.
type Location struct {
	Lat, Lon float64
	Label    string
	// PlaceID is the saved place it came from (#591), if any.
	PlaceID string
}

// locationKinds are the widget kinds that take a location.
var locationKinds = []string{"weather", "radar"}

// ErrBuiltinNotFound is returned for an unknown built-in template.
var ErrBuiltinNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The template does not exist.")

// BuiltinTemplates lists the built-in templates and creates tabs from them
// (input port).
type BuiltinTemplates interface {
	List() []BuiltinTemplate
	// CreateTab adds a tab built from the template to the account's
	// dashboard; widgets of disabled modules are left out. loc fills the
	// weather widgets; nil leaves them to be set up.
	CreateTab(ctx context.Context, id, name string, loc *Location) (TabResponse, error)
}

// ConfigMerger sets values in a widget's JSON settings (output port; JSON
// belongs to the adapters).
type ConfigMerger func(config string, values map[string]any) (string, error)

type builtinTemplates struct {
	deps  Deps
	list  []BuiltinTemplate
	merge ConfigMerger
}

// NewBuiltinTemplates returns the use case over the given templates.
func NewBuiltinTemplates(deps Deps, list []BuiltinTemplate, merge ConfigMerger) BuiltinTemplates {
	return &builtinTemplates{deps: deps, list: list, merge: merge}
}

func (uc *builtinTemplates) List() []BuiltinTemplate { return uc.list }

func (uc *builtinTemplates) CreateTab(ctx context.Context, id, name string, loc *Location) (TabResponse, error) {
	i := slices.IndexFunc(uc.list, func(t BuiltinTemplate) bool { return t.ID == id })
	if i < 0 {
		return TabResponse{}, ErrBuiltinNotFound
	}
	t := uc.list[i].Tab
	if name == "" {
		name = t.Name
	}
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return TabResponse{}, fmt.Errorf("list tabs: %w", err)
	}
	tab, err := domain.NewTab(uc.deps.IDs.NewID(), name, len(tabs))
	if err != nil {
		return TabResponse{}, err
	}
	tab = tab.WithZones(t.Zones)
	type placed struct {
		order int
		entry domain.MobileEntry
	}
	var phone []placed
	widgets := make([]domain.Widget, 0, len(t.Widgets))
	for _, cw := range t.Widgets {
		if !uc.deps.Kinds.Available(cw.Kind) {
			continue
		}
		config, err := uc.withLocation(cw.Kind, cw.Config, loc)
		if err != nil {
			return TabResponse{}, err
		}
		w, err := domain.NewWidget(uc.deps.IDs.NewID(), tab.ID(), cw.Kind, cw.Rect, config)
		if err != nil {
			return TabResponse{}, err
		}
		if err := w.Resize(domain.Size(cw.Size)); err != nil {
			return TabResponse{}, err
		}
		if err := w.MoveToZone(domain.Zone(cw.Zone)); err != nil {
			return TabResponse{}, err
		}
		widgets = append(widgets, w)
		if cw.Phone != nil {
			phone = append(phone, placed{cw.Phone.Order, domain.MobileEntry{WidgetID: w.ID(), Rows: cw.Phone.Rows, Hidden: cw.Phone.Hidden, Size: domain.Size(cw.Phone.Size)}})
		}
	}
	if len(phone) > 0 {
		slices.SortStableFunc(phone, func(a, b placed) int { return cmp.Compare(a.order, b.order) })
		entries := make([]domain.MobileEntry, 0, len(phone))
		for _, p := range phone {
			entries = append(entries, p.entry)
		}
		m, err := domain.NewMobileLayout(entries, widgets)
		if err != nil {
			return TabResponse{}, err
		}
		tab = tab.WithMobile(m)
	}
	if err := uc.deps.Templates.CreateTabWithWidgets(ctx, tab, widgets); err != nil {
		return TabResponse{}, fmt.Errorf("create tab: %w", err)
	}
	// The new tab's row takes neither the phone layout nor the zones; store
	// them next.
	if tab.Mobile() != nil || tab.Zoned() {
		if err := uc.deps.Tabs.UpdateTab(ctx, tab); err != nil {
			return TabResponse{}, fmt.Errorf("save phone layout: %w", err)
		}
	}
	return toTabResponse(ctx, tab, widgets), nil
}

// withLocation merges the location into the settings of a weather widget;
// other widgets and a missing location leave the settings as they are.
func (uc *builtinTemplates) withLocation(kind, config string, loc *Location) (string, error) {
	if loc == nil || !slices.Contains(locationKinds, kind) {
		return config, nil
	}
	if loc.Lat < -90 || loc.Lat > 90 || loc.Lon < -180 || loc.Lon > 180 {
		return "", kernel.NewValidationError("location", "The location is not valid.")
	}
	values := map[string]any{"lat": loc.Lat, "lon": loc.Lon}
	if loc.Label != "" {
		values["label"] = loc.Label
	}
	if loc.PlaceID != "" {
		values["placeId"] = loc.PlaceID
	}
	return uc.merge(config, values)
}
