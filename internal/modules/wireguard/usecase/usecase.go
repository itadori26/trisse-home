package usecase

import (
	"context"
	"fmt"
	"time"

	"trisse-home/internal/modules/wireguard/domain"
	"trisse-home/internal/shared/kernel"
)

// Export is the content of the export file.
type Export struct {
	Peers []domain.Peer
	// Names maps public keys to names found by the export script.
	Names     map[string]string
	WrittenAt time.Time
}

// PeerView is a peer at the use case boundary.
type PeerView struct {
	PublicKey, Name, Interface, Endpoint, AllowedIPs string
	Online                                           bool
	LatestHandshake                                  time.Time
	ReceivedBytes, SentBytes                         int64
}

// ListView is returned by List.
type ListView struct {
	Peers     []PeerView
	Online    int
	UpdatedAt time.Time
	Stale     bool
}

// WireGuard is the input port of the module.
type WireGuard interface {
	List(ctx context.Context) (ListView, error)
	// Rename stores a custom name for a public key; empty removes it.
	Rename(ctx context.Context, publicKey, name string) error
}

// ExportSource reads the export file (output port).
type ExportSource interface {
	Read(ctx context.Context) (Export, error)
}

// NameStore keeps names set in the UI (output port).
type NameStore interface {
	All(ctx context.Context) (map[string]string, error)
	Set(ctx context.Context, publicKey, name string) error
}

type wireguard struct {
	source ExportSource
	names  NameStore
	clock  kernel.Clock
}

// NewWireGuard returns the WireGuard use cases.
func NewWireGuard(source ExportSource, names NameStore, clock kernel.Clock) WireGuard {
	return &wireguard{source: source, names: names, clock: clock}
}

// List implements WireGuard. Names from the UI win over exported names.
func (uc *wireguard) List(ctx context.Context) (ListView, error) {
	exp, err := uc.source.Read(ctx)
	if err != nil {
		return ListView{}, err
	}
	custom, err := uc.names.All(ctx)
	if err != nil {
		return ListView{}, fmt.Errorf("load names: %w", err)
	}
	now := uc.clock.Now()
	peers := exp.Peers
	for i := range peers {
		peers[i].Name = exp.Names[peers[i].PublicKey]
		if n := custom[peers[i].PublicKey]; n != "" {
			peers[i].Name = n
		}
	}
	domain.SortForDisplay(peers, now)
	view := ListView{UpdatedAt: exp.WrittenAt, Stale: now.Sub(exp.WrittenAt) > domain.StaleAfter, Peers: make([]PeerView, 0, len(peers))}
	for _, p := range peers {
		online := p.Online(now)
		if online {
			view.Online++
		}
		view.Peers = append(view.Peers, PeerView{
			PublicKey: p.PublicKey, Name: p.DisplayName(), Interface: p.Interface, Endpoint: p.Endpoint, AllowedIPs: p.AllowedIPs,
			Online: online, LatestHandshake: p.LatestHandshake, ReceivedBytes: p.ReceivedBytes, SentBytes: p.SentBytes,
		})
	}
	return view, nil
}

// Rename implements WireGuard.
func (uc *wireguard) Rename(ctx context.Context, publicKey, name string) error {
	if !domain.ValidPublicKey(publicKey) {
		return kernel.NewValidationError("publicKey", "This is not a WireGuard public key.")
	}
	name, err := domain.NormaliseName(name)
	if err != nil {
		return err
	}
	if err := uc.names.Set(ctx, publicKey, name); err != nil {
		return fmt.Errorf("save name: %w", err)
	}
	return nil
}
