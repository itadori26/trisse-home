package wol

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

// TestRelay sends through a relay whose "broadcast" target is a local UDP
// listener, so the test needs no real broadcast.
func TestRelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	target, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	probe, _ := net.ListenPacket("udp4", "127.0.0.1:0") // reserve a free port for the relay
	relayAddr := probe.LocalAddr().String()
	_ = probe.Close()
	relay := Relay{Listen: relayAddr, Broadcast: target.LocalAddr().String(), Log: slog.New(slog.NewTextHandler(io.Discard, nil)), Burst: 1, Every: time.Hour}
	go func() { _ = relay.Serve(ctx) }()
	time.Sleep(50 * time.Millisecond)

	s := Sender{Relay: relayAddr}
	if err := s.Wake(ctx, "01:02:03:04:05:06"); err != nil {
		t.Fatalf("wake via relay: %v", err)
	}
	buf := make([]byte, 200)
	_ = target.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := target.ReadFrom(buf)
	if mac, ok := kernel.ParseMagicPacket(buf[:n]); err != nil || !ok || mac != "01:02:03:04:05:06" {
		t.Fatalf("relayed packet: %x %v", buf[:n], err)
	}
	// Burst of 1 per hour: the second packet is dropped and not acknowledged.
	if err := s.Wake(ctx, "01:02:03:04:05:06"); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("rate limit not applied: %v", err)
	}
}

func TestDirectSend(t *testing.T) {
	target, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := (Sender{Broadcast: target.LocalAddr().String()}).Wake(context.Background(), "aa:bb:cc:dd:ee:ff"); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 200)
	_ = target.SetReadDeadline(time.Now().Add(2 * time.Second))
	if n, _, err := target.ReadFrom(buf); err != nil || n != kernel.MagicPacketSize {
		t.Fatalf("direct packet: %d %v", n, err)
	}
}

func TestRelayUnreachable(t *testing.T) {
	probe, _ := net.ListenPacket("udp4", "127.0.0.1:0")
	addr := probe.LocalAddr().String()
	_ = probe.Close()
	if err := (Sender{Relay: addr}).Wake(context.Background(), "aa:bb:cc:dd:ee:ff"); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}
