package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/proxmox/domain"
	"trisse-home/internal/platform/database/dbtest"
)

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + strings.ToUpper(s), nil }
func (prefixCipher) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "enc:") {
		return "", errors.New("bad ciphertext")
	}
	return strings.ToLower(strings.TrimPrefix(s, "enc:")), nil
}

func TestSettingsRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "proxmox", Migrations())
	repo := NewSettingsRepository(db, prefixCipher{})
	if _, err := repo.Load(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	s := domain.Settings{URL: "https://pve:8006", TokenID: "a@pve!t", Secret: "uuid-secret", Fingerprint: strings.Repeat("ab", 32)}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT secret_encrypted FROM proxmox_settings`).Scan(&stored)
	if strings.Contains(stored, "uuid-secret") {
		t.Fatalf("secret stored in plain text: %q", stored)
	}
	if got, err := repo.Load(ctx); err != nil || got != s {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

const resources = `{"data":[
 {"type":"node","node":"pve","status":"online","cpu":0.25,"maxcpu":8,"mem":4000,"maxmem":16000,"disk":10,"maxdisk":100,"uptime":3600},
 {"type":"qemu","vmid":100,"name":"ha","node":"pve","status":"running","uptime":60,"cpu":0.1,"mem":1,"maxmem":2},
 {"type":"lxc","vmid":101,"name":"db","node":"pve","status":"stopped"},
 {"type":"qemu","vmid":9000,"name":"tmpl","node":"pve","status":"stopped","template":1},
 {"type":"storage","storage":"local-lvm","node":"pve","status":"available","disk":5,"maxdisk":50},
 {"type":"sdn","sdn":"localnetwork","node":"pve"}
]}`

func TestAPI(t *testing.T) {
	var posted []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "PVEAPIToken=a@pve!t=uuid" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api2/json/cluster/resources":
			_, _ = w.Write([]byte(resources))
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api2/json/nodes/"):
			posted = append(posted, r.URL.Path)
			_, _ = w.Write([]byte(`{"data":"UPID:pve:1"}`))
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	defer srv.Close()
	sum := sha256.Sum256(srv.Certificate().Raw)
	var pinnedWith []string
	pinned := func(fp string) *http.Client {
		pinnedWith = append(pinnedWith, fp)
		return srv.Client() // trusts the test certificate, like a matching pin
	}
	api := NewAPI(http.DefaultClient, pinned)
	s := domain.Settings{URL: srv.URL, TokenID: "a@pve!t", Secret: "uuid", Fingerprint: hex.EncodeToString(sum[:])}

	c, err := api.Cluster(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Nodes) != 1 || c.Nodes[0].CPUs != 8 || c.Nodes[0].Uptime != time.Hour || !c.Nodes[0].Online {
		t.Fatalf("nodes: %+v", c.Nodes)
	}
	if len(c.Guests) != 2 || c.Guests[0].Kind != domain.KindVM || c.Guests[1].Kind != domain.KindLXC {
		t.Fatalf("guests without templates: %+v", c.Guests)
	}
	if len(c.Storages) != 1 || !c.Storages[0].Active || c.Storages[0].Total != 50 {
		t.Fatalf("storages: %+v", c.Storages)
	}
	if err := api.Act(context.Background(), s, c.Guests[1], domain.ActionStart); err != nil {
		t.Fatal(err)
	}
	if len(posted) != 1 || posted[0] != "/api2/json/nodes/pve/lxc/101/status/start" {
		t.Fatalf("action path: %v", posted)
	}
	if len(pinnedWith) != 1 {
		t.Fatalf("one pinned client per fingerprint, got %d", len(pinnedWith))
	}

	s.Secret = "wrong"
	if _, err := api.Cluster(context.Background(), s); err == nil || !strings.Contains(err.Error(), "token was rejected") {
		t.Fatalf("expected token error, got %v", err)
	}
	s.Fingerprint = ""
	if _, err := api.Cluster(context.Background(), s); err == nil {
		t.Fatal("without pinning the default client must reject the self-signed certificate")
	}
}
