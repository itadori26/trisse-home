package usecase

import (
	"context"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/wireguard/domain"
)

var keyA, keyB = strings.Repeat("A", 43) + "=", strings.Repeat("B", 42) + "A="

type fakeSource struct{ exp Export }

func (f fakeSource) Read(context.Context) (Export, error) { return f.exp, nil }

type memNames map[string]string

func (m memNames) All(context.Context) (map[string]string, error) { return m, nil }
func (m memNames) Set(_ context.Context, k, v string) error {
	if v == "" {
		delete(m, k)
	} else {
		m[k] = v
	}
	return nil
}

type clk struct{ t time.Time }

func (c clk) Now() time.Time { return c.t }

func TestList(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	src := fakeSource{exp: Export{
		WrittenAt: now.Add(-10 * time.Minute),
		Names:     map[string]string{keyA: "laptop", keyB: "phone"},
		Peers: []domain.Peer{
			{PublicKey: keyA, LatestHandshake: now.Add(-time.Hour)},
			{PublicKey: keyB, LatestHandshake: now.Add(-30 * time.Second), ReceivedBytes: 42},
		},
	}}
	names := memNames{}
	uc := NewWireGuard(src, names, clk{t: now})
	if err := uc.Rename(ctx, keyA, "  Work laptop "); err != nil {
		t.Fatal(err)
	}
	if err := uc.Rename(ctx, "nope", "x"); err == nil {
		t.Fatal("invalid key accepted")
	}
	v, err := uc.List(ctx)
	if err != nil || v.Online != 1 || !v.Stale || v.Peers[0].Name != "phone" || v.Peers[1].Name != "Work laptop" {
		t.Fatalf("list %+v %v", v, err)
	}
}
