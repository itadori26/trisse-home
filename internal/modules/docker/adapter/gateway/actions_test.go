package gateway

import (
	"context"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"
)

func frame(stream byte, s string) []byte {
	h := make([]byte, 8)
	h[0] = stream
	binary.BigEndian.PutUint32(h[4:], uint32(len(s)))
	return append(h, s...)
}

func TestLogsAndRestart(t *testing.T) {
	var restarted string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/containers/homarr/logs":
			if r.URL.Query().Get("tail") != "200" {
				t.Errorf("tail = %q", r.URL.Query().Get("tail"))
			}
			body := append(frame(1, "line one\nline two\n"), frame(2, "error three\n")...)
			_, _ = w.Write(body)
		case r.Method == http.MethodGet && r.URL.Path == "/containers/tty/logs":
			_, _ = w.Write([]byte("plain\r\nlog\n"))
		case r.Method == http.MethodPost && r.URL.Path == "/containers/homarr/restart":
			restarted = "homarr"
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "forbidden", http.StatusForbidden)
		}
	}))
	defer srv.Close()
	api := NewAPI(srv.Client(), srv.URL)
	ctx := context.Background()
	lines, err := api.Logs(ctx, "homarr", 200)
	if err != nil || len(lines) != 3 || lines[2] != "error three" {
		t.Fatalf("multiplexed logs: %q %v", lines, err)
	}
	if lines, _ := api.Logs(ctx, "tty", 1); len(lines) != 1 || lines[0] != "log" {
		t.Fatalf("tty logs/limit: %q", lines)
	}
	if err := api.Restart(ctx, "homarr"); err != nil || restarted != "homarr" {
		t.Fatalf("restart: %v", err)
	}
	if err := api.Restart(ctx, "other"); err == nil {
		t.Fatal("proxy denial must be an error")
	}
}
