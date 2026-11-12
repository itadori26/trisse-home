package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 60 * time.Second
	idleTimeout       = 120 * time.Second
)

// Option customises Run.
type Option func(*options)

type options struct {
	onReady    func()
	onStopping func()
}

// WithReadyHook calls f once the listening socket is open, i.e. when requests
// can be accepted (used for systemd's READY=1).
func WithReadyHook(f func()) Option { return func(o *options) { o.onReady = f } }

// WithStoppingHook calls f when the graceful shutdown begins.
func WithStoppingHook(f func()) Option { return func(o *options) { o.onStopping = f } }

// Run serves handler on addr until ctx is cancelled, then shuts down
// gracefully within shutdownTimeout.
func Run(ctx context.Context, addr string, handler http.Handler, shutdownTimeout time.Duration, log *slog.Logger, opts ...Option) error {
	o := options{onReady: func() {}, onStopping: func() {}}
	for _, opt := range opts {
		opt(&o)
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelWarn),
	}
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", ln.Addr().String())
		errCh <- srv.Serve(ln)
	}()
	o.onReady()
	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}
	log.Info("shutting down http server")
	o.onStopping()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
