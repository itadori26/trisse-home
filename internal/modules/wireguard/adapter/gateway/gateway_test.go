package gateway

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"trisse-home/internal/modules/wireguard/domain"
	"trisse-home/internal/platform/database/dbtest"
)

var keyA, keyB = strings.Repeat("A", 43) + "=", strings.Repeat("B", 42) + "A="

func TestParseDumpIgnoresSecrets(t *testing.T) {
	dump := strings.Join([]string{
		"wg0\tPRIVATEKEYPRIVATEKEYPRIVATEKEYPRIVATEKEY0=\t" + keyA + "\t51820\toff", // interface line (5 fields)
		"wg0\t" + keyA + "\tPSKPSKPSK\t203.0.113.5:4242\t10.13.13.2/32\t1762887600\t1024\t2048\toff",
		"wg0\t" + keyB + "\t(none)\t(none)\t10.13.13.3/32\t0\t0\t0\toff",
		"garbage",
	}, "\n")
	peers, err := ParseDump(strings.NewReader(dump))
	if err != nil || len(peers) != 2 {
		t.Fatalf("%+v %v", peers, err)
	}
	if peers[0].Endpoint != "203.0.113.5:4242" || peers[0].ReceivedBytes != 1024 || peers[0].LatestHandshake.IsZero() {
		t.Fatalf("peer %+v", peers[0])
	}
	if peers[1].Endpoint != "" || !peers[1].LatestHandshake.IsZero() {
		t.Fatalf("idle peer %+v", peers[1])
	}
	for _, p := range peers {
		if strings.Contains(p.PublicKey+p.Endpoint+p.AllowedIPs, "PSK") || strings.Contains(p.PublicKey, "PRIVATE") {
			t.Fatal("secret leaked into a peer")
		}
	}
}

func TestFileSource(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	if _, err := NewFileSource(dir).Read(ctx); !errors.Is(err, domain.ErrNoExport) {
		t.Fatalf("missing export: %v", err)
	}
	_ = os.WriteFile(filepath.Join(dir, peersFile), []byte("wg0\t"+keyA+"\t-\t(none)\t10.0.0.2/32\t0\t0\t0\toff\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, namesFile), []byte(keyA+" phone\ninvalid line\n"), 0o600)
	exp, err := NewFileSource(dir).Read(ctx)
	if err != nil || len(exp.Peers) != 1 || exp.Names[keyA] != "phone" || exp.WrittenAt.IsZero() {
		t.Fatalf("%+v %v", exp, err)
	}
}

func TestNameStore(t *testing.T) {
	ctx := context.Background()
	s := NewNameStore(dbtest.Open(t, "wireguard", Migrations()))
	_ = s.Set(ctx, keyA, "laptop")
	_ = s.Set(ctx, keyA, "work laptop")
	_ = s.Set(ctx, keyB, "phone")
	_ = s.Set(ctx, keyB, "")
	all, err := s.All(ctx)
	if err != nil || len(all) != 1 || all[keyA] != "work laptop" {
		t.Fatalf("%v %v", all, err)
	}
}
