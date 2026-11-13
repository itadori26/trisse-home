package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"trisse-home/internal/modules/proxmox/domain"
	"trisse-home/internal/shared/kernel"
)

const maxResponseBytes = 4 << 20

// PinnedClientFactory returns an HTTP client that accepts exactly the
// server certificate with the given SHA-256 fingerprint (provided by the
// platform, so TLS details stay out of the adapters).
type PinnedClientFactory func(fingerprint string) *http.Client

// API talks to the Proxmox VE REST API with an API token.
type API struct {
	client *http.Client
	pinned PinnedClientFactory
	mu     sync.Mutex
	// byFingerprint caches one pinned client per fingerprint.
	byFingerprint map[string]*http.Client
}

// NewAPI returns a client; pinned may be nil when pinning is unavailable.
func NewAPI(client *http.Client, pinned PinnedClientFactory) *API {
	return &API{client: client, pinned: pinned, byFingerprint: map[string]*http.Client{}}
}

// resource is one entry of GET /cluster/resources.
type resource struct {
	Type     string  `json:"type"` // node, qemu, lxc, storage, …
	Node     string  `json:"node"`
	Status   string  `json:"status"`
	Name     string  `json:"name"`
	VMID     int     `json:"vmid"`
	Storage  string  `json:"storage"`
	Template int     `json:"template"`
	CPU      float64 `json:"cpu"`
	MaxCPU   float64 `json:"maxcpu"`
	Mem      int64   `json:"mem"`
	MaxMem   int64   `json:"maxmem"`
	Disk     int64   `json:"disk"`
	MaxDisk  int64   `json:"maxdisk"`
	Uptime   int64   `json:"uptime"`
}

// Cluster implements usecase.Client with a single request: the cluster
// resources list contains nodes, guests and storages.
func (a *API) Cluster(ctx context.Context, s domain.Settings) (domain.Cluster, error) {
	var payload struct {
		Data []resource `json:"data"`
	}
	if err := a.do(ctx, s, http.MethodGet, "/api2/json/cluster/resources", &payload); err != nil {
		return domain.Cluster{}, err
	}
	var c domain.Cluster
	for _, r := range payload.Data {
		switch r.Type {
		case "node":
			c.Nodes = append(c.Nodes, domain.Node{Name: r.Node, Online: r.Status == "online", CPU: r.CPU, CPUs: int(r.MaxCPU),
				MemUsed: r.Mem, MemTotal: r.MaxMem, DiskUsed: r.Disk, DiskTotal: r.MaxDisk, Uptime: seconds(r.Uptime)})
		case "qemu", "lxc":
			if r.Template == 1 {
				continue // templates cannot run
			}
			c.Guests = append(c.Guests, domain.Guest{ID: r.VMID, Name: r.Name, Node: r.Node, Kind: domain.GuestKind(r.Type),
				Status: r.Status, Uptime: seconds(r.Uptime), CPU: r.CPU, MemUsed: r.Mem, MemTotal: r.MaxMem})
		case "storage":
			c.Storages = append(c.Storages, domain.Storage{Name: r.Storage, Node: r.Node, Used: r.Disk, Total: r.MaxDisk,
				Active: r.Status == "available"})
		}
	}
	return c, nil
}

// Act implements usecase.Client: POST …/status/start or …/status/shutdown.
func (a *API) Act(ctx context.Context, s domain.Settings, g domain.Guest, act domain.Action) error {
	path := fmt.Sprintf("/api2/json/nodes/%s/%s/%s/status/%s",
		url.PathEscape(g.Node), url.PathEscape(string(g.Kind)), strconv.Itoa(g.ID), url.PathEscape(string(act)))
	var payload struct {
		Data string `json:"data"` // the task ID (UPID)
	}
	return a.do(ctx, s, http.MethodPost, path, &payload)
}

func seconds(s int64) time.Duration { return time.Duration(s) * time.Second }

func (a *API) clientFor(s domain.Settings) *http.Client {
	if s.Fingerprint == "" || a.pinned == nil {
		return a.client
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	c, ok := a.byFingerprint[s.Fingerprint]
	if !ok {
		c = a.pinned(s.Fingerprint)
		a.byFingerprint[s.Fingerprint] = c
	}
	return c
}

func (a *API) do(ctx context.Context, s domain.Settings, method, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, method, s.URL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "PVEAPIToken="+s.TokenID+"="+s.Secret)
	res, err := a.clientFor(s).Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: the API token was rejected (%s)", kernel.ErrTargetAuth, res.Status)
	case http.StatusForbidden:
		return fmt.Errorf("%w: the API token lacks a permission (%s)", kernel.ErrTargetAuth, res.Status)
	default:
		return fmt.Errorf("unexpected answer from Proxmox VE: %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from Proxmox VE: %w", err)
	}
	return nil
}
