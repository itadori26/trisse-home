package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchSetupStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    string
		wantErr bool
	}{
		{"no account yet", http.StatusOK, `{"setupRequired":true,"authenticated":false}`, setupPending, false},
		{"account exists", http.StatusOK, `{"setupRequired":false,"authenticated":false}`, setupDone, false},
		{"field missing", http.StatusOK, `{"authenticated":false}`, "", true},
		{"not json", http.StatusOK, `<html>`, "", true},
		{"server error", http.StatusInternalServerError, `{}`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/auth/status" {
					http.NotFound(w, r)
					return
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()
			got, err := fetchSetupStatus(context.Background(), srv.Client(), srv.URL)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("got %q, %v; want %q, error %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestFetchSetupStatusUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	if _, err := fetchSetupStatus(context.Background(), http.DefaultClient, url); err == nil {
		t.Fatal("expected an error for an unreachable server")
	}
}

func TestLoopbackURL(t *testing.T) {
	for in, want := range map[string]string{"": "http://127.0.0.1:8090", ":80": "http://127.0.0.1:80", "0.0.0.0:9000": "http://127.0.0.1:9000"} {
		got, err := loopbackURL(in)
		if err != nil || got != want {
			t.Fatalf("loopbackURL(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := loopbackURL("8090"); err == nil {
		t.Fatal("expected an error for an address without port")
	}
}
