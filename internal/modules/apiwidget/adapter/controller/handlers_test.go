package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/apiwidget/domain"
	"trisse-home/internal/modules/apiwidget/usecase"
	"trisse-home/internal/shared/kernel"
)

type fakeUC struct {
	usecase.APIWidget
	got domain.EndpointInput
	id  string
}

func (f *fakeUC) Endpoints(context.Context) ([]domain.Endpoint, error) {
	return []domain.Endpoint{{ID: "a", Name: "NAS", URL: "http://nas/api", HeaderName: "X-Api-Key", HeaderValue: "s3cret", Interval: time.Minute}}, nil
}

func (f *fakeUC) UpdateEndpoint(_ context.Context, id string, in domain.EndpointInput) (domain.Endpoint, error) {
	f.id, f.got = id, in
	return domain.Endpoint{ID: id}, nil
}

func (f *fakeUC) FetchValues(_ context.Context, id string) (usecase.Values, error) {
	if id == "missing" {
		return usecase.Values{}, domain.ErrNotFound
	}
	return usecase.Values{Error: kernel.NewPublicErrorf(kernel.ErrInvalid, "The endpoint answered {status}.", "status", "401 Unauthorized")}, nil
}

func TestHandlers(t *testing.T) {
	uc := &fakeUC{}
	h := NewHandlers(uc)

	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/apiwidget/endpoints", nil))
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "s3cret") || !strings.Contains(rec.Body.String(), `"hasHeader":true`) {
		t.Fatalf("list: %d %s", rec.Code, rec.Body)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/apiwidget/endpoints/a", strings.NewReader(
		`{"name":"NAS","url":"http://nas/api","headerName":"X-Api-Key","keepHeader":true,"intervalSeconds":90,
		  "extractions":[{"label":"Temp","path":"$.t","warn":70,"crit":null,"below":true}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "a")
	rec = httptest.NewRecorder()
	h.Update(rec, req)
	x := uc.got.Extractions
	if rec.Code != http.StatusOK || uc.id != "a" || uc.got.Interval != 90*time.Second || !uc.got.KeepHeader ||
		len(x) != 1 || *x[0].Warn != 70 || x[0].Crit != nil || !x[0].Below {
		t.Fatalf("update: %d %+v", rec.Code, uc.got)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/apiwidget/endpoints/a/values", nil)
	req.SetPathValue("id", "a")
	rec = httptest.NewRecorder()
	h.Values(rec, req)
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `"errorKey":"`+kernel.MessageKey("The endpoint answered {status}.")+`"`) ||
		!strings.Contains(body, `"status":"401 Unauthorized"`) {
		t.Fatalf("values: %d %s", rec.Code, body)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/apiwidget/endpoints/missing/values", nil)
	req.SetPathValue("id", "missing")
	rec = httptest.NewRecorder()
	h.Values(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing: %d", rec.Code)
	}
}
