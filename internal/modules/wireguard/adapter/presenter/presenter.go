// Package presenter maps WireGuard peers to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"trisse-home/internal/modules/wireguard/usecase"
)

// PeerDTO is one peer; latestHandshake is null when there never was one.
type PeerDTO struct {
	PublicKey       string     `json:"publicKey"`
	Name            string     `json:"name"`
	Interface       string     `json:"interface"`
	Endpoint        string     `json:"endpoint"`
	AllowedIPs      string     `json:"allowedIps"`
	Online          bool       `json:"online"`
	LatestHandshake *time.Time `json:"latestHandshake"`
	ReceivedBytes   int64      `json:"receivedBytes"`
	SentBytes       int64      `json:"sentBytes"`
}

// ListDTO is returned by GET /api/wireguard/peers.
type ListDTO struct {
	Peers     []PeerDTO `json:"peers"`
	Online    int       `json:"online"`
	UpdatedAt time.Time `json:"updatedAt"`
	Stale     bool      `json:"stale"`
}

// List maps a ListView.
func List(v usecase.ListView) ListDTO {
	dto := ListDTO{Online: v.Online, UpdatedAt: v.UpdatedAt, Stale: v.Stale, Peers: make([]PeerDTO, 0, len(v.Peers))}
	for _, p := range v.Peers {
		d := PeerDTO{PublicKey: p.PublicKey, Name: p.Name, Interface: p.Interface, Endpoint: p.Endpoint, AllowedIPs: p.AllowedIPs,
			Online: p.Online, ReceivedBytes: p.ReceivedBytes, SentBytes: p.SentBytes}
		if !p.LatestHandshake.IsZero() {
			t := p.LatestHandshake
			d.LatestHandshake = &t
		}
		dto.Peers = append(dto.Peers, d)
	}
	return dto
}
