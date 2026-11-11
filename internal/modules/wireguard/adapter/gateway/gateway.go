// Package gateway reads the WireGuard export written by
// scripts/wireguard-export.sh and stores peer names in SQLite.
//
// Ring: Interface Adapters.
package gateway

import (
	"bufio"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/wireguard/domain"
	"trisse-home/internal/modules/wireguard/usecase"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// File names inside the export directory.
const (
	peersFile = "peers.dump"
	namesFile = "peer-names.txt"
	maxBytes  = 1 << 20
)

// FileSource implements usecase.ExportSource for an export directory.
type FileSource struct{ dir string }

// NewFileSource reads exports from dir (empty disables the module's data).
func NewFileSource(dir string) *FileSource { return &FileSource{dir: dir} }

// Read implements usecase.ExportSource.
func (f *FileSource) Read(_ context.Context) (usecase.Export, error) {
	if f.dir == "" {
		return usecase.Export{}, domain.ErrNoExport
	}
	file, err := os.Open(filepath.Join(f.dir, peersFile))
	if errors.Is(err, fs.ErrNotExist) {
		return usecase.Export{}, domain.ErrNoExport
	}
	if err != nil {
		return usecase.Export{}, fmt.Errorf("open export: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return usecase.Export{}, fmt.Errorf("stat export: %w", err)
	}
	peers, err := ParseDump(io.LimitReader(file, maxBytes))
	if err != nil {
		return usecase.Export{}, err
	}
	return usecase.Export{Peers: peers, Names: f.names(), WrittenAt: info.ModTime()}, nil
}

func (f *FileSource) names() map[string]string {
	out := map[string]string{}
	file, err := os.Open(filepath.Join(f.dir, namesFile))
	if err != nil {
		return out // names are optional
	}
	defer file.Close()
	sc := bufio.NewScanner(io.LimitReader(file, maxBytes))
	for sc.Scan() {
		key, name, ok := strings.Cut(strings.TrimSpace(sc.Text()), " ")
		if ok && domain.ValidPublicKey(key) {
			if n, err := domain.NormaliseName(name); err == nil {
				out[key] = n
			}
		}
	}
	return out
}

// ParseDump parses peer lines of `wg show all dump` (9 tab-separated fields:
// interface, public key, preshared key, endpoint, allowed IPs, latest
// handshake, rx, tx, keepalive). Every other line — in particular the
// interface line carrying the private key — is ignored, and the preshared
// key column is never read.
func ParseDump(r io.Reader) ([]domain.Peer, error) {
	var peers []domain.Peer
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		if len(f) != 9 || !domain.ValidPublicKey(f[1]) {
			continue
		}
		p := domain.Peer{Interface: f[0], PublicKey: f[1], AllowedIPs: f[4]}
		if f[3] != "(none)" {
			p.Endpoint = f[3]
		}
		if ts, _ := strconv.ParseInt(f[5], 10, 64); ts > 0 {
			p.LatestHandshake = time.Unix(ts, 0).UTC()
		}
		p.ReceivedBytes, _ = strconv.ParseInt(f[6], 10, 64)
		p.SentBytes, _ = strconv.ParseInt(f[7], 10, 64)
		peers = append(peers, p)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read export: %w", err)
	}
	return peers, nil
}

// NameStore implements usecase.NameStore.
type NameStore struct{ db *sql.DB }

// NewNameStore returns a store backed by db.
func NewNameStore(db *sql.DB) *NameStore { return &NameStore{db: db} }

// All implements usecase.NameStore.
func (s *NameStore) All(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT public_key, name FROM wireguard_peer_names`)
	if err != nil {
		return nil, fmt.Errorf("query names: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, n string
		if err := rows.Scan(&k, &n); err != nil {
			return nil, fmt.Errorf("scan name: %w", err)
		}
		out[k] = n
	}
	return out, rows.Err()
}

// Set implements usecase.NameStore; an empty name deletes the entry.
func (s *NameStore) Set(ctx context.Context, publicKey, name string) error {
	var err error
	if name == "" {
		_, err = s.db.ExecContext(ctx, `DELETE FROM wireguard_peer_names WHERE public_key = ?`, publicKey)
	} else {
		_, err = s.db.ExecContext(ctx, `INSERT INTO wireguard_peer_names (public_key, name) VALUES (?, ?)
			ON CONFLICT (public_key) DO UPDATE SET name = excluded.name`, publicKey, name)
	}
	if err != nil {
		return fmt.Errorf("store name: %w", err)
	}
	return nil
}
