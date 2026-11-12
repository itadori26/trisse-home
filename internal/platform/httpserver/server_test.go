package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRunServesAndShutsDownGracefully(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	go func() { done <- Run(ctx, addr, handler, time.Second, slog.New(slog.NewTextHandler(io.Discard, nil))) }()

	var res *http.Response
	for range 50 {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+addr, nil)
		if res, err = http.DefaultClient.Do(req); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("server not reachable: %v", err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusTeapot {
		t.Fatalf("status %d", res.StatusCode)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("shutdown returned %v", err)
	}
}

func TestRunReportsListenErrors(t *testing.T) {
	err := Run(context.Background(), "256.0.0.1:1", http.NotFoundHandler(), time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err == nil {
		t.Fatal("expected listen error")
	}
}

func TestRunCallsHooksWhenListeningAndStopping(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan struct{})
	stopping := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, "127.0.0.1:0", http.NotFoundHandler(), time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)),
			WithReadyHook(func() { close(ready) }),
			WithStoppingHook(func() { close(stopping) }))
	}()
	select {
	case <-ready:
	case err := <-done:
		t.Fatalf("server stopped before it was ready: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("ready hook not called")
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("shutdown returned %v", err)
	}
	select {
	case <-stopping:
	default:
		t.Fatal("stopping hook not called")
	}
}

func TestRunDoesNotReportReadyOnListenError(t *testing.T) {
	called := false
	err := Run(context.Background(), "256.0.0.1:1", http.NotFoundHandler(), time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)),
		WithReadyHook(func() { called = true }))
	if err == nil || called {
		t.Fatalf("err = %v, ready called = %v", err, called)
	}
}
