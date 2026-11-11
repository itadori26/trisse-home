package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"trisse-home/internal/modules/services/domain"
	"trisse-home/internal/shared/kernel"
	"trisse-home/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Repository implements usecase.Repository on SQLite.
type Repository struct{ db *sql.DB }

// NewRepository returns a repository backed by db.
func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

const columns = `id, name, url, check_url, icon, description, skip_tls_verify, position, mac`

func scan(s interface{ Scan(...any) error }) (domain.Service, error) {
	var svc domain.Service
	var skip int
	var mac string
	err := s.Scan(&svc.ID, &svc.Name, &svc.URL, &svc.CheckURL, &svc.Icon, &svc.Description, &skip, &svc.Position, &mac)
	svc.SkipTLSVerify, svc.MAC = skip == 1, kernel.MAC(mac)
	return svc, err
}

// List returns all services ordered by position.
func (r *Repository) List(ctx context.Context) ([]domain.Service, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+columns+` FROM services ORDER BY position, name`)
	if err != nil {
		return nil, fmt.Errorf("query services: %w", err)
	}
	defer rows.Close()
	var out []domain.Service
	for rows.Next() {
		s, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Get returns one service.
func (r *Repository) Get(ctx context.Context, id string) (domain.Service, error) {
	s, err := scan(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM services WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Service{}, domain.ErrServiceNotFound
	}
	if err != nil {
		return domain.Service{}, fmt.Errorf("select service: %w", err)
	}
	return s, nil
}

// Create inserts a service.
func (r *Repository) Create(ctx context.Context, s domain.Service) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO services (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.URL, s.CheckURL, s.Icon, s.Description, boolInt(s.SkipTLSVerify), s.Position, string(s.MAC))
	if err != nil {
		return fmt.Errorf("insert service: %w", err)
	}
	return nil
}

// Update stores all fields of a service.
func (r *Repository) Update(ctx context.Context, s domain.Service) error {
	_, err := r.db.ExecContext(ctx, `UPDATE services SET name = ?, url = ?, check_url = ?, icon = ?, description = ?,
		skip_tls_verify = ?, position = ?, mac = ? WHERE id = ?`,
		s.Name, s.URL, s.CheckURL, s.Icon, s.Description, boolInt(s.SkipTLSVerify), s.Position, string(s.MAC), s.ID)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	return nil
}

// Delete removes a service.
func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM services WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// StatusStore is an in-memory usecase.StatusStore.
type StatusStore struct {
	mu sync.RWMutex
	m  map[string]domain.Status
}

// NewStatusStore returns an empty store.
func NewStatusStore() *StatusStore { return &StatusStore{m: map[string]domain.Status{}} }

// Get returns the latest status of a service.
func (s *StatusStore) Get(id string) (domain.Status, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.m[id]
	return st, ok
}

// Put stores a status.
func (s *StatusStore) Put(st domain.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[st.ServiceID] = st
}

// Delete forgets the status of a service.
func (s *StatusStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id)
}

// HTTPChecker implements usecase.HealthChecker with two clients: one that
// verifies TLS certificates and one that does not (opt-in per service).
type HTTPChecker struct {
	verifying *http.Client
	insecure  *http.Client
}

// NewHTTPChecker returns a checker using the given clients.
func NewHTTPChecker(verifying, insecure *http.Client) *HTTPChecker {
	return &HTTPChecker{verifying: verifying, insecure: insecure}
}

// maxDrainBytes limits how much of a response body is read before closing.
const maxDrainBytes = 4096

// Check issues a GET request and reports status code, latency and the
// expiry of the server's TLS certificate.
func (c *HTTPChecker) Check(ctx context.Context, url string, skipTLSVerify bool) domain.CheckResult {
	client := c.verifying
	if skipTLSVerify {
		client = c.insecure
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.CheckResult{Err: fmt.Errorf("build request: %w", err)}
	}
	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return domain.CheckResult{Err: err}
	}
	result := domain.CheckResult{HTTPStatus: res.StatusCode, Latency: time.Since(start)}
	if res.TLS != nil && len(res.TLS.PeerCertificates) > 0 {
		result.CertExpiresAt = res.TLS.PeerCertificates[0].NotAfter
	}
	_, _ = io.CopyN(io.Discard, res.Body, maxDrainBytes)
	_ = res.Body.Close()
	return result
}

// HistoryStore keeps health check results in SQLite.
type HistoryStore struct{ db *sql.DB }

// NewHistoryStore returns a store backed by db.
func NewHistoryStore(db *sql.DB) *HistoryStore { return &HistoryStore{db: db} }

// Add implements usecase.HistoryStore.
func (h *HistoryStore) Add(ctx context.Context, p domain.LatencyPoint) error {
	_, err := h.db.ExecContext(ctx, `INSERT OR REPLACE INTO services_history (service_id, at, latency_ms, online) VALUES (?, ?, ?, ?)`,
		p.ServiceID, sqlutil.FormatTime(p.At), p.LatencyMs, sqlutil.BoolToInt(p.Online))
	if err != nil {
		return fmt.Errorf("insert history: %w", err)
	}
	return nil
}

// Since implements usecase.HistoryStore; points are ordered by time.
func (h *HistoryStore) Since(ctx context.Context, t time.Time) ([]domain.LatencyPoint, error) {
	rows, err := h.db.QueryContext(ctx, `SELECT service_id, at, latency_ms, online FROM services_history WHERE at >= ? ORDER BY at`, sqlutil.FormatTime(t))
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()
	var out []domain.LatencyPoint
	for rows.Next() {
		var p domain.LatencyPoint
		var at string
		var online int
		if err := rows.Scan(&p.ServiceID, &at, &p.LatencyMs, &online); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		if p.At, err = sqlutil.ParseTime(at); err != nil {
			return nil, err
		}
		p.Online = online == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

// PruneBefore implements usecase.HistoryStore.
func (h *HistoryStore) PruneBefore(ctx context.Context, t time.Time) error {
	if _, err := h.db.ExecContext(ctx, `DELETE FROM services_history WHERE at < ?`, sqlutil.FormatTime(t)); err != nil {
		return fmt.Errorf("prune history: %w", err)
	}
	return nil
}
