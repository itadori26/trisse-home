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

const columns = `id, name, url, check_url, icon, description, skip_tls_verify, position`

func scan(s interface{ Scan(...any) error }) (domain.Service, error) {
	var svc domain.Service
	var skip int
	err := s.Scan(&svc.ID, &svc.Name, &svc.URL, &svc.CheckURL, &svc.Icon, &svc.Description, &skip, &svc.Position)
	svc.SkipTLSVerify = skip == 1
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
	_, err := r.db.ExecContext(ctx, `INSERT INTO services (`+columns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.URL, s.CheckURL, s.Icon, s.Description, boolInt(s.SkipTLSVerify), s.Position)
	if err != nil {
		return fmt.Errorf("insert service: %w", err)
	}
	return nil
}

// Update stores all fields of a service.
func (r *Repository) Update(ctx context.Context, s domain.Service) error {
	_, err := r.db.ExecContext(ctx, `UPDATE services SET name = ?, url = ?, check_url = ?, icon = ?, description = ?,
		skip_tls_verify = ?, position = ? WHERE id = ?`,
		s.Name, s.URL, s.CheckURL, s.Icon, s.Description, boolInt(s.SkipTLSVerify), s.Position, s.ID)
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

// Check issues a GET request and reports status code and latency.
func (c *HTTPChecker) Check(ctx context.Context, url string, skipTLSVerify bool) (int, time.Duration, error) {
	client := c.verifying
	if skipTLSVerify {
		client = c.insecure
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("build request: %w", err)
	}
	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	latency := time.Since(start)
	_, _ = io.CopyN(io.Discard, res.Body, maxDrainBytes)
	_ = res.Body.Close()
	return res.StatusCode, latency, nil
}
