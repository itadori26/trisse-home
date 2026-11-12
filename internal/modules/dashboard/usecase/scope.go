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
