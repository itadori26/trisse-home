package domain

import (
	"strings"
	"testing"
	"time"
)

func TestPeers(t *testing.T) {
	now := time.Date(2026, 11, 11, 20, 0, 0, 0, time.UTC)
	key := strings.Repeat("A", 43) + "="
	if !ValidPublicKey(key) || ValidPublicKey("short=") {
		t.Fatal("public key validation wrong")
	}
	peers := []Peer{
		{PublicKey: key, Name: "old", LatestHandshake: now.Add(-time.Hour)},
		{PublicKey: key, Name: "phone", LatestHandshake: now.Add(-time.Minute)},
		{PublicKey: "BBBBBBBBBBBB", LatestHandshake: time.Time{}},
	}
	if !peers[1].Online(now) || peers[0].Online(now) || peers[2].Online(now) {
		t.Fatal("online rule wrong")
	}
	SortForDisplay(peers, now)
	if peers[0].Name != "phone" || peers[2].DisplayName() != "BBBBBBBB…" {
		t.Fatalf("order %+v", peers)
	}
	if _, err := NormaliseName(strings.Repeat("x", 41)); err == nil {
		t.Fatal("long name accepted")
	}
}
