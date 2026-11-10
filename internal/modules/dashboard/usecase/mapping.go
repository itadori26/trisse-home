package usecase

import "trisse-home/internal/modules/dashboard/domain"

func toWidgetResponse(w domain.Widget) WidgetResponse {
	r := w.Rect()
	return WidgetResponse{ID: w.ID(), TabID: w.TabID(), Kind: w.Kind(), X: r.X, Y: r.Y, W: r.W, H: r.H, Config: w.Config()}
}

func toTabResponse(t domain.Tab, widgets []domain.Widget) TabResponse {
	res := TabResponse{ID: t.ID(), Name: t.Name(), Position: t.Position(), Widgets: []WidgetResponse{}}
	for _, w := range widgets {
		if w.TabID() == t.ID() {
			res.Widgets = append(res.Widgets, toWidgetResponse(w))
		}
	}
	return res
}
