package controller

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/modules/backup/usecase"
)

type fakeRestorer struct {
	req     usecase.RestoreRequest
	content string
}

func (f *fakeRestorer) Restore(_ context.Context, req usecase.RestoreRequest) error {
	f.req = req
	if req.Upload != nil {
		b, _ := io.ReadAll(req.Upload)
		f.content = string(b)
	}
	return nil
}

func TestRestoreHandler(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("password", "secret")
	_ = mw.WriteField("code", " 123456 ")
	_ = mw.WriteField("source", "upload")
	fw, _ := mw.CreateFormFile("file", "backup.db")
	_, _ = fw.Write([]byte("sqlite-bytes"))
	_ = mw.Close()

	uc := &fakeRestorer{}
	req := httptest.NewRequest(http.MethodPost, "/api/backups/restore", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	NewRestoreHandlers(uc, 1<<20).Restore(rec, req)
	if rec.Code != http.StatusAccepted || uc.req.Password != "secret" || uc.req.Code != "123456" || uc.req.Source != usecase.FromUpload || uc.content != "sqlite-bytes" {
		t.Fatalf("%d %s %+v %q", rec.Code, rec.Body, uc.req, uc.content)
	}

	rec = httptest.NewRecorder()
	NewRestoreHandlers(uc, 1<<20).Restore(rec, httptest.NewRequest(http.MethodPost, "/api/backups/restore", bytes.NewReader([]byte(`{}`))))
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("json body: %d", rec.Code)
	}
}
