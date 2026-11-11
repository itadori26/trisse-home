package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/dashboard/domain"
)

type manageTemplates struct{ deps Deps }

// NewManageTemplates returns the ManageTemplates use case.
func NewManageTemplates(deps Deps) ManageTemplates { return &manageTemplates{deps: deps} }

// List implements ManageTemplates.
func (uc *manageTemplates) List(ctx context.Context) ([]TemplateResponse, error) {
	ts, err := uc.deps.Templates.ListTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	out := make([]TemplateResponse, 0, len(ts))
	for _, t := range ts {
		out = append(out, toTemplateResponse(t))
	}
	return out, nil
}

// SaveTab implements ManageTemplates.
func (uc *manageTemplates) SaveTab(ctx context.Context, tabID, name string) (TemplateResponse, error) {
	if _, err := uc.deps.Tabs.GetTab(ctx, tabID); err != nil {
		return TemplateResponse{}, err
	}
	existing, err := uc.deps.Templates.ListTemplates(ctx)
	if err != nil {
		return TemplateResponse{}, fmt.Errorf("list templates: %w", err)
	}
	if len(existing) >= domain.MaxTemplates {
		return TemplateResponse{}, domain.ErrTooManyTemplates
	}
	widgets, err := uc.deps.Widgets.ListWidgetsByTab(ctx, tabID)
	if err != nil {
		return TemplateResponse{}, fmt.Errorf("list widgets: %w", err)
	}
	t, err := domain.NewTemplate(uc.deps.IDs.NewID(), name, widgets)
	if err != nil {
		return TemplateResponse{}, err
	}
	if err := uc.deps.Templates.CreateTemplate(ctx, t); err != nil {
		return TemplateResponse{}, fmt.Errorf("create template: %w", err)
	}
	return toTemplateResponse(t), nil
}

// Delete implements ManageTemplates.
func (uc *manageTemplates) Delete(ctx context.Context, id string) error {
	if _, err := uc.deps.Templates.GetTemplate(ctx, id); err != nil {
		return err
	}
	return uc.deps.Templates.DeleteTemplate(ctx, id)
}

// CreateTab implements ManageTemplates. Widgets of kinds that are no longer
// available (disabled modules) are skipped.
func (uc *manageTemplates) CreateTab(ctx context.Context, templateID, name string) (TabResponse, error) {
	t, err := uc.deps.Templates.GetTemplate(ctx, templateID)
	if err != nil {
		return TabResponse{}, err
	}
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
	widgets := make([]domain.Widget, 0, len(t.Widgets))
	for _, tw := range t.Widgets {
		if !uc.deps.Kinds.Available(tw.Kind) {
			continue
		}
		w, err := domain.NewWidget(uc.deps.IDs.NewID(), tab.ID(), tw.Kind, tw.Rect, tw.Config)
		if err != nil {
			continue // a template written by an older version must not block the rest
		}
		widgets = append(widgets, w)
	}
	if err := uc.deps.Templates.CreateTabWithWidgets(ctx, tab, widgets); err != nil {
		return TabResponse{}, fmt.Errorf("create tab: %w", err)
	}
	return toTabResponse(tab, widgets), nil
}
