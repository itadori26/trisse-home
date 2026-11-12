package main

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"trisse-home/internal/platform/config"
)

func (c *client) restore(fields map[string]string, file []byte, wantStatus int) []byte {
	c.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for _, k := range []string{"password", "code", "source", "name"} {
		if v, ok := fields[k]; ok {
			_ = mw.WriteField(k, v)
		}
	}
	if file != nil {
		fw, _ := mw.CreateFormFile("file", "backup.db")
		_, _ = fw.Write(file)
	}
	_ = mw.Close()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/backups/restore", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != wantStatus {
		c.t.Fatalf("restore: status %d, want %d: %s", res.StatusCode, wantStatus, out)
	}
	return out
}

func TestRestoreEndToEnd(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "restore.db")
	srv := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) { c.DatabasePath = dbPath })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf, _ = session["csrfToken"].(string)
	var backupFile obj
	c.json(http.MethodPost, "/api/backups", nil, http.StatusCreated, &backupFile)
	name := backupFile["name"].(string)

	c.restore(map[string]string{"password": "wrong-password", "source": "local", "name": name}, nil, http.StatusUnauthorized)
	c.restore(map[string]string{"password": "e2e-password-123", "source": "upload"}, []byte("not a database, just text that is long enough"), http.StatusUnprocessableEntity)
	if _, err := os.Stat(dbPath + ".restore"); err == nil {
		t.Fatal("rejected restore must not be scheduled")
	}

	// Drain a restart signal left by other tests.
	select {
	case <-restartRequests:
	default:
	}
	c.restore(map[string]string{"password": "e2e-password-123", "source": "local", "name": name}, nil, http.StatusAccepted)
	if _, err := os.Stat(dbPath + ".restore"); err != nil {
		t.Fatalf("restore not scheduled: %v", err)
	}
	select {
	case <-restartRequests:
	default:
		t.Fatal("no restart requested")
	}
	entries, _ := os.ReadDir(filepath.Dir(dbPath))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".restore-upload-") {
			t.Fatalf("staging file left: %s", e.Name())
		}
	}
}
