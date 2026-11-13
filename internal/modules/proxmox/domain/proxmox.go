package domain

import (
	"encoding/hex"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// ErrNotConfigured is returned while no Proxmox VE host is configured.
var ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Proxmox VE is not configured yet.")

// tokenIDPattern matches API token IDs such as monitor@pve!trisse-home.
var tokenIDPattern = regexp.MustCompile(`^[^\s@!]+@[^\s@!]+![A-Za-z][A-Za-z0-9._-]*$`)

// Settings are the connection details of a Proxmox VE host or cluster.
type Settings struct {
	URL string
	// TokenID is the API token ID (user@realm!name), Secret its UUID.
	TokenID string
	Secret  string
	// Fingerprint pins the server certificate (SHA-256 of the leaf, hex)
	// instead of verifying it against the system CAs; Proxmox uses a
	// self-signed certificate by default.
	Fingerprint string
}

// NewSettings validates and normalises the settings.
func NewSettings(rawURL, tokenID, secret, fingerprint string) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || (u.Path != "" && u.Path != "/") {
		return Settings{}, kernel.NewValidationError("url", "The URL must look like https://192.168.1.10:8006.")
	}
	tokenID = strings.TrimSpace(tokenID)
	if !tokenIDPattern.MatchString(tokenID) {
		return Settings{}, kernel.NewValidationError("tokenId", "The token ID must look like monitor@pve!trisse-home.")
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return Settings{}, kernel.NewValidationError("secret", "The token secret is required.")
	}
	fp, err := NormaliseFingerprint(fingerprint)
	if err != nil {
		return Settings{}, err
	}
	if fp != "" && u.Scheme != "https" {
		return Settings{}, kernel.NewValidationError("fingerprint", "A certificate fingerprint needs an https:// URL.")
	}
	return Settings{URL: rawURL, TokenID: tokenID, Secret: secret, Fingerprint: fp}, nil
}

// NormaliseFingerprint accepts a SHA-256 fingerprint with or without
// colons (as shown by Proxmox under Certificates) and returns lower-case hex.
func NormaliseFingerprint(raw string) (string, error) {
	fp := strings.ToLower(strings.NewReplacer(":", "", " ", "").Replace(strings.TrimSpace(raw)))
	if fp == "" {
		return "", nil
	}
	if b, err := hex.DecodeString(fp); err != nil || len(b) != 32 {
		return "", kernel.NewValidationError("fingerprint", "The fingerprint must be a SHA-256 value such as AB:CD:…, 32 bytes.")
	}
	return fp, nil
}

// Node is a Proxmox VE cluster node.
type Node struct {
	Name   string
	Online bool
	// CPU is the load between 0 and 1 over CPUs cores.
	CPU                 float64
	CPUs                int
	MemUsed, MemTotal   int64
	DiskUsed, DiskTotal int64
	Uptime              time.Duration
}

// GuestKind distinguishes virtual machines and containers.
type GuestKind string

// Guest kinds, named like the Proxmox API path segments.
const (
	KindVM  GuestKind = "qemu"
	KindLXC GuestKind = "lxc"
)

// Guest is a VM or LXC container.
type Guest struct {
	ID     int
	Name   string
	Node   string
	Kind   GuestKind
	Status string // running, stopped, paused, …
	Uptime time.Duration
	CPU    float64
	// MemUsed and MemTotal are in bytes.
	MemUsed, MemTotal int64
}

// Running reports whether the guest is running.
func (g Guest) Running() bool { return g.Status == "running" }

// Storage is a storage of a node.
type Storage struct {
	Name, Node  string
	Used, Total int64
	Active      bool
}

// Cluster is a snapshot of the whole cluster.
type Cluster struct {
	Nodes    []Node
	Guests   []Guest
	Storages []Storage
}

// Sort orders nodes and storages by name and guests by ID, the order the
// Proxmox UI uses.
func (c *Cluster) Sort() {
	slices.SortFunc(c.Nodes, func(a, b Node) int { return strings.Compare(a.Name, b.Name) })
	slices.SortFunc(c.Guests, func(a, b Guest) int { return a.ID - b.ID })
	slices.SortFunc(c.Storages, func(a, b Storage) int {
		if c := strings.Compare(a.Node, b.Node); c != 0 {
			return c
		}
		return strings.Compare(a.Name, b.Name)
	})
}

// Guest returns the guest with the given ID.
func (c Cluster) Guest(id int) (Guest, bool) {
	i := slices.IndexFunc(c.Guests, func(g Guest) bool { return g.ID == id })
	if i < 0 {
		return Guest{}, false
	}
	return c.Guests[i], true
}

// Action is a power action on a guest.
type Action string

// Supported actions: a clean shutdown, never a hard stop.
const (
	ActionStart    Action = "start"
	ActionShutdown Action = "shutdown"
)

// ErrActionNotAllowed is returned when the guest is not on the allow-list.
var ErrActionNotAllowed = kernel.NewPublicError(kernel.ErrForbidden,
	"Actions are not enabled for this guest (PROXMOX_ACTIONS_GUESTS).")

// ErrUnknownAction is returned for anything but start and shutdown.
var ErrUnknownAction = kernel.NewPublicError(kernel.ErrInvalid, "Unknown action, use start or shutdown.")

// ErrGuestNotFound is returned for an unknown guest ID.
var ErrGuestNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The guest does not exist.")

// ActionPolicy lists the guests that may be started and shut down. Actions
// are off while the list is empty; it comes from the environment, so a
// session alone can never widen it.
type ActionPolicy struct {
	Allowed []int
}

// Allows reports whether actions are permitted for the guest ID.
func (p ActionPolicy) Allows(id int) bool { return slices.Contains(p.Allowed, id) }

// Check validates an action request.
func (p ActionPolicy) Check(id int, a Action) error {
	if a != ActionStart && a != ActionShutdown {
		return ErrUnknownAction
	}
	if !p.Allows(id) {
		return ErrActionNotAllowed
	}
	return nil
}
