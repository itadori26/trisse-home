package wol

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Defaults.
const (
	DefaultBroadcast = "255.255.255.255:9"
	DefaultRelayPort = "9009"
	ackTimeout       = 2 * time.Second
	ack              = "ok"
)

// ErrRelayUnreachable is returned when the relay did not acknowledge.
var ErrRelayUnreachable = kernel.NewPublicError(kernel.ErrUnavailable,
	"The Wake-on-LAN relay did not answer. Is COMPOSE_PROFILES=wol set (see README)?")

// sendBroadcast sends p to a broadcast address.
func sendBroadcast(ctx context.Context, addr string, p []byte) error {
	lc := net.ListenConfig{Control: allowBroadcast}
	conn, err := lc.ListenPacket(ctx, "udp4", ":0")
	if err != nil {
		return fmt.Errorf("open udp socket: %w", err)
	}
	defer conn.Close()
	dst, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", addr, err)
	}
	if _, err := conn.WriteTo(p, dst); err != nil {
		return fmt.Errorf("send magic packet: %w", err)
	}
	return nil
}

// Sender implements the services module's Waker port.
type Sender struct {
	// Relay is host:port of the relay; empty sends directly to Broadcast.
	Relay     string
	Broadcast string
}

// Wake sends the magic packet for mac.
func (s Sender) Wake(ctx context.Context, mac kernel.MAC) error {
	p := kernel.MagicPacket(mac)
	if s.Relay == "" {
		return sendBroadcast(ctx, s.Broadcast, p)
	}
	var d net.Dialer
	conn, err := d.DialContext(ctx, "udp", s.Relay)
	if err != nil {
		return fmt.Errorf("dial relay: %w", err)
	}
	defer conn.Close()
	if _, err := conn.Write(p); err != nil {
		return fmt.Errorf("send to relay: %w", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(ackTimeout))
	buf := make([]byte, 8)
	n, err := conn.Read(buf)
	if err != nil || string(buf[:n]) != ack {
		return ErrRelayUnreachable
	}
	return nil
}

// Relay receives magic packets and re-broadcasts them into the LAN.
type Relay struct {
	Listen    string
	Broadcast string
	Log       *slog.Logger
	// Burst and Every limit the packet rate (token bucket).
	Burst int
	Every time.Duration
}

// Serve runs the relay until ctx is cancelled.
func (r Relay) Serve(ctx context.Context) error {
	var lc net.ListenConfig
	conn, err := lc.ListenPacket(ctx, "udp", r.Listen)
	if err != nil {
		return fmt.Errorf("listen %s: %w", r.Listen, err)
	}
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()
	r.Log.Info("wake-on-lan relay listening", "addr", conn.LocalAddr().String(), "broadcast", r.Broadcast)
	limit := newBucket(r.Burst, r.Every)
	buf := make([]byte, 512)
	for {
		n, from, err := conn.ReadFrom(buf)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("read: %w", err)
		}
		mac, ok := kernel.ParseMagicPacket(buf[:n])
		if !ok {
			continue // not a magic packet: ignore silently
		}
		if !limit.take(time.Now()) {
			r.Log.Warn("wake-on-lan rate limit exceeded", "from", from.String())
			continue
		}
		if err := sendBroadcast(ctx, r.Broadcast, buf[:n]); err != nil {
			r.Log.Error("wake-on-lan broadcast failed", "err", err)
			continue
		}
		r.Log.Info("magic packet relayed", "mac", string(mac), "from", from.String())
		_, _ = conn.WriteTo([]byte(ack), from)
	}
}

// bucket is a minimal token bucket.
type bucket struct {
	mu     sync.Mutex
	tokens float64
	burst  float64
	every  time.Duration
	last   time.Time
}

func newBucket(burst int, every time.Duration) *bucket {
	if burst < 1 {
		burst = 10
	}
	if every <= 0 {
		every = time.Second
	}
	return &bucket{tokens: float64(burst), burst: float64(burst), every: every}
}

func (b *bucket) take(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.last.IsZero() {
		b.tokens = min(b.burst, b.tokens+float64(now.Sub(b.last))/float64(b.every))
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}
