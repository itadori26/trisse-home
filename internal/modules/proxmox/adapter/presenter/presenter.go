// Package presenter maps Proxmox use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"math"
	"slices"
	"time"

	"trisse-home/internal/modules/proxmox/usecase"
)

// SettingsDTO is returned by GET/PUT /api/proxmox/settings; the token
// secret is never returned.
type SettingsDTO struct {
	Configured  bool   `json:"configured"`
	URL         string `json:"url"`
	TokenID     string `json:"tokenId"`
	HasSecret   bool   `json:"hasSecret"`
	Fingerprint string `json:"fingerprint"`
}

// NodeDTO is a cluster node; CPU is a percentage, sizes are bytes.
type NodeDTO struct {
	Name          string  `json:"name"`
	Online        bool    `json:"online"`
	CPUPercent    float64 `json:"cpuPercent"`
	CPUs          int     `json:"cpus"`
	MemUsed       int64   `json:"memUsed"`
	MemTotal      int64   `json:"memTotal"`
	DiskUsed      int64   `json:"diskUsed"`
	DiskTotal     int64   `json:"diskTotal"`
	UptimeSeconds int64   `json:"uptimeSeconds"`
}

// GuestDTO is a VM or LXC container.
type GuestDTO struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Node          string  `json:"node"`
	Kind          string  `json:"kind" tstype:"'qemu' | 'lxc'"`
	Status        string  `json:"status"`
	UptimeSeconds int64   `json:"uptimeSeconds"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemUsed       int64   `json:"memUsed"`
	MemTotal      int64   `json:"memTotal"`
	// Actions is true when start/shutdown are allowed (PROXMOX_ACTIONS_GUESTS).
	Actions bool `json:"actions"`
}

// StorageDTO is a storage of a node.
type StorageDTO struct {
	Name   string `json:"name"`
	Node   string `json:"node"`
	Used   int64  `json:"used"`
	Total  int64  `json:"total"`
	Active bool   `json:"active"`
}

// StatusDTO is returned by GET /api/proxmox/status.
type StatusDTO struct {
	Nodes     []NodeDTO    `json:"nodes"`
	Guests    []GuestDTO   `json:"guests"`
	Storages  []StorageDTO `json:"storages"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Error     string       `json:"error"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

func percent(f float64) float64 { return math.Round(f*1000) / 10 }

// Status maps a StatusView.
func Status(v usecase.StatusView) StatusDTO {
	c := v.Cluster
	dto := StatusDTO{UpdatedAt: v.UpdatedAt, Error: v.Error, Nodes: make([]NodeDTO, 0, len(c.Nodes)),
		Guests: make([]GuestDTO, 0, len(c.Guests)), Storages: make([]StorageDTO, 0, len(c.Storages))}
	for _, n := range c.Nodes {
		dto.Nodes = append(dto.Nodes, NodeDTO{Name: n.Name, Online: n.Online, CPUPercent: percent(n.CPU), CPUs: n.CPUs,
			MemUsed: n.MemUsed, MemTotal: n.MemTotal, DiskUsed: n.DiskUsed, DiskTotal: n.DiskTotal, UptimeSeconds: int64(n.Uptime.Seconds())})
	}
	for _, g := range c.Guests {
		dto.Guests = append(dto.Guests, GuestDTO{ID: g.ID, Name: g.Name, Node: g.Node, Kind: string(g.Kind), Status: g.Status,
			UptimeSeconds: int64(g.Uptime.Seconds()), CPUPercent: percent(g.CPU), MemUsed: g.MemUsed, MemTotal: g.MemTotal,
			Actions: slices.Contains(v.Actions, g.ID)})
	}
	for _, s := range c.Storages {
		dto.Storages = append(dto.Storages, StorageDTO(s))
	}
	return dto
}
