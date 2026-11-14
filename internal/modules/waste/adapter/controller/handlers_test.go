package controller

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/waste/domain"
	"trisse-home/internal/modules/waste/usecase"
)

type fakeUC struct {
	usecase.Waste
	uploaded *[]byte
}

func (f fakeUC) UploadCalendar(_ context.Context, name string, content []byte) (domain.Settings, error) {
	*f.uploaded = content
	day := time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC)
	src, err := domain.NewFileSource(name, day, []domain.Collection{{Date: day}})
	return domain.Settings{Source: src, Rules: []domain.Rule{}}, err
}

func (fakeUC) Upcoming(context.Context) (usecase.Overview, error) {
	return usecase.Overview{}, domain.ErrNotConfigured
}

func (fakeUC) Settings(context.Context) (domain.Settings, bool, error) {
	return domain.Settings{Rules: []domain.Rule{}}, false, nil
}

func TestHandlers(t *testing.T) {
	h := NewHandlers(fakeUC{uploaded: new([]byte)})
	rec := httptest.NewRecorder()
	h.Upcoming(rec, httptest.NewRequest(http.MethodGet, "/api/waste/upcoming", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("upcoming: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.Settings(rec, httptest.NewRequest(http.MethodGet, "/api/waste/settings", nil))
	if rec.Code != http.StatusOK || rec.Body.String() != `{"configured":false,"source":"","url":"","file":null,"rules":[]}`+"\n" {
		t.Fatalf("settings: %d %s", rec.Code, rec.Body)
	}
}

func upload(t *testing.T, h *Handlers, name, contentType string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
	header.Set("Content-Type", contentType)
	part, err := mw.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(content)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/waste/calendar", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	h.UploadCalendar(rec, req)
	return rec
}

func TestUploadCalendar(t *testing.T) {
	uploaded := new([]byte)
	h := NewHandlers(fakeUC{uploaded: uploaded})
	ics := []byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n")

	rec := upload(t, h, "abfall-2027.ics", "application/octet-stream", ics)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"source":"file"`) ||
		!strings.Contains(rec.Body.String(), `"name":"abfall-2027.ics"`) || !bytes.Equal(*uploaded, ics) {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body)
	}
	if rec := upload(t, h, "plan", "text/calendar; charset=utf-8", ics); rec.Code != http.StatusOK {
		t.Fatalf("text/calendar without .ics: %d", rec.Code)
	}
	if rec := upload(t, h, "plan.pdf", "application/pdf", []byte("%PDF")); rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("wrong type: %d", rec.Code)
	}
	big := bytes.Repeat([]byte("x"), domain.MaxCalendarBytes+1)
	if rec := upload(t, h, "big.ics", "text/calendar", big); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize: %d", rec.Code)
	}
	huge := bytes.Repeat([]byte("x"), domain.MaxCalendarBytes+2*uploadOverhead)
	if rec := upload(t, h, "huge.ics", "text/calendar", huge); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("body limit: %d", rec.Code)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/waste/calendar", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.UploadCalendar(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("json body: %d", rec.Code)
	}
}
