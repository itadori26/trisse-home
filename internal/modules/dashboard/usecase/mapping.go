package usecase

import (
	"context"
	"slices"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
)

func toWidgetResponse(w domain.Widget) WidgetResponse {
	r := w.Rect()
	return WidgetResponse{ID: w.ID(), TabID: w.TabID(), Kind: w.Kind(), X: r.X, Y: r.Y, W: r.W, H: r.H, Config: w.Config(), Size: string(w.Size()), Zone: string(w.Zone())}
}

func toTabResponse(ctx context.Context, t domain.Tab, widgets []domain.Widget) TabResponse {
	res := TabResponse{ID: t.ID(), Name: t.Name(), Position: t.Position(), Background: t.Background(), Zoned: t.Zoned(), Sample: t.Sample(), Widgets: []WidgetResponse{}, Rules: []RuleView{},
		Mine: t.Access().Owner == "" || t.Access().Mine(kernel.PrincipalFrom(ctx)), Visibility: string(t.Access().Visibility)}
	for _, r := range t.Rules() {
		from, to := r.Range()
		res.Rules = append(res.Rules, RuleView{Days: r.Days(), From: from, To: to, Networks: r.Networks()})
	}
	var own []domain.Widget
	for _, w := range widgets {
		if w.TabID() == t.ID() {
			own = append(own, w)
			res.Widgets = append(res.Widgets, toWidgetResponse(w))
		}
	}
	mobile := domain.ResolveMobileLayout(t.Mobile(), own)
	res.MobileCustom = mobile.Custom()
	res.Mobile = []MobileView{}
	for _, e := range mobile.Entries() {
		res.Mobile = append(res.Mobile, MobileView{WidgetID: e.WidgetID, Hidden: e.Hidden, Rows: e.Rows, Size: string(e.Size)})
	}
	return res
}

func toTemplateResponse(t domain.Template) TemplateResponse {
	res := TemplateResponse{ID: t.ID, Name: t.Name, Widgets: len(t.Widgets), Kinds: []string{}}
	for _, w := range t.Widgets {
		if !slices.Contains(res.Kinds, w.Kind) {
			res.Kinds = append(res.Kinds, w.Kind)
		}
	}
	return res
}
