package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/briefing/domain"
	"trisse-home/internal/modules/briefing/usecase"
)

type fakeBuild struct{}

func (fakeBuild) Execute(context.Context) (usecase.BriefingView, error) {
	return usecase.BriefingView{At: time.Date(2026, 11, 12, 7, 0, 0, 0, time.UTC), Unavailable: []domain.SectionKind{domain.SectionEvents},
		Briefing: domain.Briefing{Sections: []domain.Section{{Kind: domain.SectionTodos, More: 1,
			Items: []domain.Item{{Text: "Rent", Detail: "overdue", Severity: domain.SeverityWarning, At: time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)}}}}}}, nil
}

type fakeSettings struct{ got usecase.SettingsInput }

func (f *fakeSettings) Get(context.Context) (usecase.SettingsView, error) {
	return usecase.SettingsView{}, nil
}
func (f *fakeSettings) Save(_ context.Context, in usecase.SettingsInput) (usecase.SettingsView, error) {
	f.got = in
	return usecase.SettingsView{Layout: in.Layout, PushAt: in.PushAt}, nil
}

type fakeSend struct{ sent bool }

func (f *fakeSend) Execute(context.Context) (bool, error) { return false, nil }
func (f *fakeSend) SendNow(context.Context) error         { f.sent = true; return nil }

func TestHandlers(t *testing.T) {
	settings, send := &fakeSettings{}, &fakeSend{}
	h := NewHandlers(UseCases{Build: fakeBuild{}, Settings: settings, Send: send})

	rec := httptest.NewRecorder()
	h.Get(rec, httptest.NewRequest(http.MethodGet, "/api/briefing", nil))
	want := `"sections":[{"kind":"todos","items":[{"text":"Rent","detail":"overdue","at":"2026-11-10T00:00:00Z","severity":"warning"}],"more":1}],"unavailable":["events"]`
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), want) {
		t.Fatalf("get %d %s", rec.Code, rec.Body)
	}

	rec = httptest.NewRecorder()
	h.SaveSettings(rec, httptest.NewRequest(http.MethodPut, "/api/briefing/settings", strings.NewReader(
		`{"layout":[{"kind":"waste","enabled":false}],"pushOn":true,"pushAt":"06:30","pushDays":[1],"skipEmpty":true,"lat":1,"lon":2}`)))
	if rec.Code != http.StatusOK || len(settings.got.Layout) != 1 || settings.got.Layout[0].Kind != domain.SectionWaste ||
		settings.got.PushAt != "06:30" || settings.got.Lon != 2 {
		t.Fatalf("save %d %+v", rec.Code, settings.got)
	}
	rec = httptest.NewRecorder()
	h.SaveSettings(rec, httptest.NewRequest(http.MethodPut, "/api/briefing/settings", strings.NewReader(`{"nope":1}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown field %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.Send(rec, httptest.NewRequest(http.MethodPost, "/api/briefing/send", nil))
	if rec.Code != http.StatusNoContent || !send.sent {
		t.Fatalf("send %d", rec.Code)
	}
}
