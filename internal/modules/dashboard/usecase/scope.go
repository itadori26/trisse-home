package usecase

import (
	"context"
	"fmt"
	"slices"
)

// TabScope answers which widget kinds are placed on a set of tabs; the
// auth module uses it to decide what a guest may read (input port).
type TabScope interface {
	WidgetKindsOnTabs(ctx context.Context, tabIDs []string) ([]string, error)
	// TabIDs lists every tab in display order (the display of an appliance
	// shows them all, #140).
	TabIDs(ctx context.Context) ([]string, error)
}

type tabScope struct{ deps Deps }

// NewTabScope returns the TabScope use case.
func NewTabScope(deps Deps) TabScope { return &tabScope{deps: deps} }

// WidgetKindsOnTabs implements TabScope; the result is sorted and unique.
func (uc *tabScope) WidgetKindsOnTabs(ctx context.Context, tabIDs []string) ([]string, error) {
	widgets, err := uc.deps.Widgets.ListWidgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list widgets: %w", err)
	}
	var kinds []string
	for _, w := range widgets {
		if slices.Contains(tabIDs, w.TabID()) && !slices.Contains(kinds, w.Kind()) {
			kinds = append(kinds, w.Kind())
		}
	}
	slices.Sort(kinds)
	return kinds, nil
}

// TabIDs implements TabScope.
func (uc *tabScope) TabIDs(ctx context.Context) ([]string, error) {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tabs: %w", err)
	}
	ids := make([]string, 0, len(tabs))
	for _, t := range tabs {
		ids = append(ids, t.ID())
	}
	return ids, nil
}
