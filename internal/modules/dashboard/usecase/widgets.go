package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/dashboard/domain"
)

type manageWidgets struct{ deps Deps }

// NewManageWidgets returns the ManageWidgets use case.
func NewManageWidgets(deps Deps) ManageWidgets { return &manageWidgets{deps: deps} }

// Add implements ManageWidgets.
func (uc *manageWidgets) Add(ctx context.Context, req AddWidgetRequest) (WidgetResponse, error) {
	if !uc.deps.Kinds.Available(req.Kind) {
		return WidgetResponse{}, domain.ErrUnknownWidgetKind
	}
	if _, err := uc.deps.Tabs.GetTab(ctx, req.TabID); err != nil {
		return WidgetResponse{}, err
	}
	rect, err := domain.NewRect(req.X, req.Y, req.W, req.H)
	if err != nil {
		return WidgetResponse{}, err
	}
	w, err := domain.NewWidget(uc.deps.IDs.NewID(), req.TabID, req.Kind, rect, req.Config)
	if err != nil {
		return WidgetResponse{}, err
	}
	if err := uc.deps.Widgets.CreateWidget(ctx, w); err != nil {
		return WidgetResponse{}, fmt.Errorf("create widget: %w", err)
	}
	return toWidgetResponse(w), nil
}

// Configure implements ManageWidgets.
func (uc *manageWidgets) Configure(ctx context.Context, req ConfigureWidgetRequest) (WidgetResponse, error) {
	w, err := uc.deps.Widgets.GetWidget(ctx, req.WidgetID)
	if err != nil {
		return WidgetResponse{}, err
	}
	if err := w.Configure(req.Config); err != nil {
		return WidgetResponse{}, err
	}
	if err := uc.deps.Widgets.UpdateWidget(ctx, w); err != nil {
		return WidgetResponse{}, fmt.Errorf("update widget: %w", err)
	}
	return toWidgetResponse(w), nil
}

// SaveLayout implements ManageWidgets. Placements of widgets that do not
// belong to the tab are rejected so a stale client cannot move foreign widgets.
func (uc *manageWidgets) SaveLayout(ctx context.Context, req SaveLayoutRequest) error {
	widgets, err := uc.deps.Widgets.ListWidgetsByTab(ctx, req.TabID)
	if err != nil {
		return fmt.Errorf("list widgets: %w", err)
	}
	byID := make(map[string]domain.Widget, len(widgets))
	for _, w := range widgets {
		byID[w.ID()] = w
	}
	changed := make([]domain.Widget, 0, len(req.Placements))
	for _, p := range req.Placements {
		w, ok := byID[p.WidgetID]
		if !ok {
			return domain.ErrWidgetNotFound
		}
		rect, err := domain.NewRect(p.X, p.Y, p.W, p.H)
		if err != nil {
			return err
		}
		w.Place(rect)
		changed = append(changed, w)
	}
	if err := uc.deps.Widgets.SaveWidgetRects(ctx, changed); err != nil {
		return fmt.Errorf("save layout: %w", err)
	}
	return nil
}

// Delete implements ManageWidgets.
func (uc *manageWidgets) Delete(ctx context.Context, widgetID string) error {
	if _, err := uc.deps.Widgets.GetWidget(ctx, widgetID); err != nil {
		return err
	}
	if err := uc.deps.Widgets.DeleteWidget(ctx, widgetID); err != nil {
		return fmt.Errorf("delete widget: %w", err)
	}
	return nil
}
