package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/dashboard/domain"
)

type manageTabs struct{ deps Deps }

// NewManageTabs returns the ManageTabs use case.
func NewManageTabs(deps Deps) ManageTabs { return &manageTabs{deps: deps} }

// Create implements ManageTabs; new tabs are appended at the end.
func (uc *manageTabs) Create(ctx context.Context, req CreateTabRequest) (TabResponse, error) {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return TabResponse{}, fmt.Errorf("list tabs: %w", err)
	}
	tab, err := domain.NewTab(uc.deps.IDs.NewID(), req.Name, len(tabs))
	if err != nil {
		return TabResponse{}, err
	}
	if err := uc.deps.Tabs.CreateTab(ctx, tab); err != nil {
		return TabResponse{}, fmt.Errorf("create tab: %w", err)
	}
	return toTabResponse(tab, nil), nil
}

// Rename implements ManageTabs.
func (uc *manageTabs) Rename(ctx context.Context, req RenameTabRequest) (TabResponse, error) {
	tab, err := uc.deps.Tabs.GetTab(ctx, req.TabID)
	if err != nil {
		return TabResponse{}, err
	}
	if err := tab.Rename(req.Name); err != nil {
		return TabResponse{}, err
	}
	if err := uc.deps.Tabs.UpdateTab(ctx, tab); err != nil {
		return TabResponse{}, fmt.Errorf("update tab: %w", err)
	}
	widgets, err := uc.deps.Widgets.ListWidgetsByTab(ctx, tab.ID())
	if err != nil {
		return TabResponse{}, fmt.Errorf("list widgets: %w", err)
	}
	return toTabResponse(tab, widgets), nil
}

// Delete implements ManageTabs. The last tab is kept so the dashboard is
// never empty.
func (uc *manageTabs) Delete(ctx context.Context, tabID string) error {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return fmt.Errorf("list tabs: %w", err)
	}
	remaining := make([]domain.Tab, 0, len(tabs))
	found := false
	for _, t := range tabs {
		if t.ID() == tabID {
			found = true
			continue
		}
		remaining = append(remaining, t)
	}
	if !found {
		return domain.ErrTabNotFound
	}
	if len(remaining) == 0 {
		return domain.ErrLastTab
	}
	if err := uc.deps.Tabs.DeleteTab(ctx, tabID); err != nil {
		return fmt.Errorf("delete tab: %w", err)
	}
	return uc.persistOrder(ctx, remaining)
}

// Reorder implements ManageTabs.
func (uc *manageTabs) Reorder(ctx context.Context, req ReorderTabsRequest) error {
	tabs, err := uc.deps.Tabs.ListTabs(ctx)
	if err != nil {
		return fmt.Errorf("list tabs: %w", err)
	}
	byID := make(map[string]domain.Tab, len(tabs))
	for _, t := range tabs {
		byID[t.ID()] = t
	}
	if len(req.TabIDs) != len(tabs) {
		return domain.ErrTabOrderMismatch
	}
	ordered := make([]domain.Tab, 0, len(tabs))
	for _, id := range req.TabIDs {
		t, ok := byID[id]
		if !ok {
			return domain.ErrTabOrderMismatch
		}
		delete(byID, id)
		ordered = append(ordered, t)
	}
	return uc.persistOrder(ctx, ordered)
}

func (uc *manageTabs) persistOrder(ctx context.Context, tabs []domain.Tab) error {
	for i := range tabs {
		tabs[i].MoveTo(i)
	}
	if err := uc.deps.Tabs.SaveTabOrder(ctx, tabs); err != nil {
		return fmt.Errorf("save tab order: %w", err)
	}
	return nil
}

// SetBackground implements ManageTabs.
// SetRules implements ManageTabs.
func (uc *manageTabs) SetRules(ctx context.Context, tabID string, in []domain.RuleInput) (TabResponse, error) {
	rules, err := domain.NewTabRules(in)
	if err != nil {
		return TabResponse{}, err
	}
	tab, err := uc.deps.Tabs.GetTab(ctx, tabID)
	if err != nil {
		return TabResponse{}, err
	}
	tab = tab.WithRules(rules)
	if err := uc.deps.Tabs.UpdateTab(ctx, tab); err != nil {
		return TabResponse{}, fmt.Errorf("update tab: %w", err)
	}
	widgets, err := uc.deps.Widgets.ListWidgetsByTab(ctx, tab.ID())
	if err != nil {
		return TabResponse{}, fmt.Errorf("list widgets: %w", err)
	}
	return toTabResponse(tab, widgets), nil
}

func (uc *manageTabs) SetBackground(ctx context.Context, tabID, background string) (TabResponse, error) {
	tab, err := uc.deps.Tabs.GetTab(ctx, tabID)
	if err != nil {
		return TabResponse{}, err
	}
	if err := tab.SetBackground(background); err != nil {
		return TabResponse{}, err
	}
	if err := uc.deps.Tabs.UpdateTab(ctx, tab); err != nil {
		return TabResponse{}, fmt.Errorf("update tab: %w", err)
	}
	widgets, err := uc.deps.Widgets.ListWidgetsByTab(ctx, tab.ID())
	if err != nil {
		return TabResponse{}, fmt.Errorf("list widgets: %w", err)
	}
	return toTabResponse(tab, widgets), nil
}
