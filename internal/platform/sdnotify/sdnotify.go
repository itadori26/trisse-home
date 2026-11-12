package sdnotify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"
)

// States understood by the service manager (see sd_notify(3)).
const (
	stateReady    = "READY=1"
	stateStopping = "STOPPING=1"
	stateWatchdog = "WATCHDOG=1"
)

// Notifier sends state changes to the service manager. The zero value and a
// nil *Notifier are valid no-ops, which is what FromEnv returns when the
// process does not run under systemd (Docker, development).
type Notifier struct {
	socket   string
	watchdog time.Duration
}

// FromEnv reads NOTIFY_SOCKET and the watchdog settings (WATCHDOG_USEC,
// WATCHDOG_PID) through getenv. pid is the current process ID; the watchdog
// is only enabled when WATCHDOG_PID is unset or names this process, so that
// child processes never feed the watchdog of their parent.
func FromEnv(getenv func(string) string, pid int) *Notifier {
	n := &Notifier{socket: getenv("NOTIFY_SOCKET")}
	if n.socket == "" {
		return n
	}
	if p := getenv("WATCHDOG_PID"); p != "" && p != strconv.Itoa(pid) {
		return n
	}
	if usec, err := strconv.ParseInt(getenv("WATCHDOG_USEC"), 10, 64); err == nil && usec > 0 {
		n.watchdog = time.Duration(usec) * time.Microsecond
	}
	return n
}

// Enabled reports whether a service manager listens for notifications.
func (n *Notifier) Enabled() bool { return n != nil && n.socket != "" }

// WatchdogInterval returns the timeout configured with WatchdogSec= and
// whether the watchdog is enabled.
func (n *Notifier) WatchdogInterval() (time.Duration, bool) {
	if !n.Enabled() || n.watchdog <= 0 {
		return 0, false
	}
	return n.watchdog, true
}

// Ready tells the service manager that start-up is complete.
func (n *Notifier) Ready() error { return n.send(stateReady) }

// Stopping tells the service manager that a graceful shutdown has begun.
func (n *Notifier) Stopping() error { return n.send(stateStopping) }

// Watchdog sends a keep-alive ping.
func (n *Notifier) Watchdog() error { return n.send(stateWatchdog) }

func (n *Notifier) send(state string) error {
	if !n.Enabled() {
		return nil
	}
	addr := &net.UnixAddr{Name: n.socket, Net: "unixgram"}
	if strings.HasPrefix(addr.Name, "@") {
		// Abstract socket namespace (Linux): the leading @ stands for a NUL byte.
		addr.Name = "\x00" + addr.Name[1:]
	}
	conn, err := net.DialUnix("unixgram", nil, addr)
	if err != nil {
		return fmt.Errorf("sd_notify %s: %w", state, err)
	}
	defer conn.Close()
	if _, err := conn.Write([]byte(state)); err != nil {
		return fmt.Errorf("sd_notify %s: %w", state, err)
	}
	return nil
}

// HealthCheck reports whether the service still works. The watchdog only pings
// while it succeeds, so a hung or broken process is restarted by systemd.
type HealthCheck func(ctx context.Context) error

// RunWatchdog pings the service manager every half watchdog interval as long as
// check succeeds, until ctx is cancelled. It returns immediately when the
// watchdog is disabled. Each check gets a quarter of the interval as timeout,
// so a slow check cannot delay the next ping beyond the deadline.
func RunWatchdog(ctx context.Context, n *Notifier, check HealthCheck, log *slog.Logger) {
	interval, ok := n.WatchdogInterval()
	if !ok {
		return
	}
	ticker := time.NewTicker(interval / 2)
	defer ticker.Stop()
	for {
		ping(ctx, n, check, interval/4, log)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func ping(ctx context.Context, n *Notifier, check HealthCheck, timeout time.Duration, log *slog.Logger) {
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if err := check(checkCtx); err != nil {
		if !errors.Is(ctx.Err(), context.Canceled) {
			log.Warn("health check failed, withholding watchdog ping", "err", err)
		}
		return
	}
	if err := n.Watchdog(); err != nil {
		log.Warn("watchdog ping failed", "err", err)
	}
}
