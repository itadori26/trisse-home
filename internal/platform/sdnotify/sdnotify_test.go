package sdnotify

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

// listen opens a datagram socket that plays the service manager.
func listen(t *testing.T, name string) *net.UnixConn {
	t.Helper()
	conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: name, Net: "unixgram"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func receive(t *testing.T, conn *net.UnixConn) string {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("no notification received: %v", err)
	}
	return string(buf[:n])
}

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestFromEnv(t *testing.T) {
	pid := 4242
	tests := []struct {
		name     string
		env      map[string]string
		enabled  bool
		watchdog time.Duration
	}{
		{"no systemd", map[string]string{}, false, 0},
		{"notify only", map[string]string{"NOTIFY_SOCKET": "/run/x"}, true, 0},
		{"watchdog", map[string]string{"NOTIFY_SOCKET": "/run/x", "WATCHDOG_USEC": "60000000"}, true, time.Minute},
		{"watchdog for this pid", map[string]string{"NOTIFY_SOCKET": "/run/x", "WATCHDOG_USEC": "1000", "WATCHDOG_PID": strconv.Itoa(pid)}, true, time.Millisecond},
		{"watchdog for another pid", map[string]string{"NOTIFY_SOCKET": "/run/x", "WATCHDOG_USEC": "1000", "WATCHDOG_PID": "1"}, true, 0},
		{"invalid usec", map[string]string{"NOTIFY_SOCKET": "/run/x", "WATCHDOG_USEC": "soon"}, true, 0},
		{"watchdog without socket", map[string]string{"WATCHDOG_USEC": "1000"}, false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := FromEnv(env(tt.env), pid)
			if n.Enabled() != tt.enabled {
				t.Fatalf("Enabled() = %v, want %v", n.Enabled(), tt.enabled)
			}
			got, ok := n.WatchdogInterval()
			if got != tt.watchdog || ok != (tt.watchdog > 0) {
				t.Fatalf("WatchdogInterval() = %v, %v; want %v", got, ok, tt.watchdog)
			}
		})
	}
}

func TestDisabledNotifierIsANoOp(t *testing.T) {
	var nilNotifier *Notifier
	for _, n := range []*Notifier{nilNotifier, {}, FromEnv(env(nil), 1)} {
		if err := n.Ready(); err != nil {
			t.Fatal(err)
		}
		if err := n.Watchdog(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSendsStates(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "notify")
	conn := listen(t, socket)
	n := FromEnv(env(map[string]string{"NOTIFY_SOCKET": socket}), 1)

	for _, tc := range []struct {
		send func() error
		want string
	}{
		{n.Ready, "READY=1"},
		{n.Watchdog, "WATCHDOG=1"},
		{n.Stopping, "STOPPING=1"},
	} {
		if err := tc.send(); err != nil {
			t.Fatal(err)
		}
		if got := receive(t, conn); got != tc.want {
			t.Fatalf("got %q, want %q", got, tc.want)
		}
	}
}

func TestAbstractSocket(t *testing.T) {
	name := "trisse-home-test-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	conn := listen(t, "\x00"+name)
	n := FromEnv(env(map[string]string{"NOTIFY_SOCKET": "@" + name}), 1)
	if err := n.Ready(); err != nil {
		t.Fatal(err)
	}
	if got := receive(t, conn); got != "READY=1" {
		t.Fatalf("got %q", got)
	}
}

func TestSendReportsMissingSocket(t *testing.T) {
	n := FromEnv(env(map[string]string{"NOTIFY_SOCKET": filepath.Join(t.TempDir(), "missing")}), 1)
	if err := n.Ready(); err == nil {
		t.Fatal("expected an error for a missing socket")
	}
}

func TestRunWatchdogPingsOnlyWhileHealthy(t *testing.T) {
	socket := filepath.Join(t.TempDir(), "notify")
	conn := listen(t, socket)
	n := FromEnv(env(map[string]string{"NOTIFY_SOCKET": socket, "WATCHDOG_USEC": "40000"}), 1) // 40 ms

	var healthy atomic.Bool
	healthy.Store(true)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		RunWatchdog(ctx, n, func(context.Context) error {
			if healthy.Load() {
				return nil
			}
			return errors.New("database unavailable")
		}, discard)
		close(done)
	}()

	if got := receive(t, conn); got != "WATCHDOG=1" {
		t.Fatalf("got %q", got)
	}
	healthy.Store(false)
	// Drain a ping that may already be in flight, then expect silence.
	_ = conn.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
	_, _ = conn.Read(make([]byte, 64))
	_ = conn.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	if _, err := conn.Read(make([]byte, 64)); err == nil {
		t.Fatal("watchdog pinged although the health check fails")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunWatchdog did not stop after cancel")
	}
}

func TestRunWatchdogReturnsWhenDisabled(t *testing.T) {
	done := make(chan struct{})
	go func() {
		RunWatchdog(context.Background(), FromEnv(env(nil), 1), func(context.Context) error { return nil }, discard)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunWatchdog must return immediately without a watchdog")
	}
}
