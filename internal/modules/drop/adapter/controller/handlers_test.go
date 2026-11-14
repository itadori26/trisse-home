package controller

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/drop/domain"
	"trisse-home/internal/modules/drop/usecase"
)

type fakeUC struct {
	usecase.Drop
	expires time.Duration
	data    []byte
}

func (f *fakeUC) AddFile(_ context.Context, name, ct string, data []byte, expiresIn time.Duration) (domain.Item, error) {
	f.expires, f.data = expiresIn, data
	return domain.NewFile("f", name, ct, int64(len(data)), expiresIn, time.Now())
}

func (f *fakeUC) File(context.Context, string) (domain.Item, []byte, error) {
	return domain.Item{Kind: domain.KindFile, FileName: "Bericht ä.html", ContentType: "text/html"}, []byte("<script>"), nil
}

func TestHandlers(t *testing.T) {
	uc := &fakeUC{}
	h := NewHandlers(uc)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("expiresInHours", "168")
	fw, _ := mw.CreateFormFile("file", "notes.txt")
	_, _ = fw.Write([]byte("hello"))
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/drop/files", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	h.AddFile(rec, req)
	if rec.Code != http.StatusCreated || uc.expires != 168*time.Hour || string(uc.data) != "hello" || !strings.Contains(rec.Body.String(), `"fileName":"notes.txt"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}

	rec = httptest.NewRecorder()
	h.File(rec, httptest.NewRequest(http.MethodGet, "/api/drop/f/file", nil))
	hd := rec.Header()
	if hd.Get("Content-Type") != "application/octet-stream" || !strings.HasPrefix(hd.Get("Content-Disposition"), "attachment") ||
		!strings.Contains(hd.Get("Content-Disposition"), "filename*=UTF-8''Bericht%20%C3%A4.html") || hd.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("an HTML upload must be a download: %v", hd)
	}
}
