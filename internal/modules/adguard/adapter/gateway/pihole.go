package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	"trisse-home/internal/modules/adguard/domain"
)

// PiHole queries the Pi-hole v6 REST API (#83).
//
// Pi-hole limits the number of concurrent sessions, so one session per
// connection is kept and reused; a new one is only requested when Pi-hole
// answers 401 (expired or revoked). Sessions live in memory only.
type PiHole struct {
	client *http.Client
	mu     sync.Mutex
	// sessions maps a connection (URL and password) to its session ID.
	sessions map[string]string
}

// NewPiHole returns a Pi-hole client.
func NewPiHole(client *http.Client) *PiHole {
	return &PiHole{client: client, sessions: map[string]string{}}
}

// errPassword is reported when the login fails.
var errPassword = errors.New("the password was rejected by Pi-hole")

// errUnauthorized marks a 401 answer that a fresh session may fix.
var errUnauthorized = errors.New("unauthorized: Pi-hole answered 401")

type piholeSummary struct {
	Queries struct {
		Total   int `json:"total"`
		Blocked int `json:"blocked"`
	} `json:"queries"`
}

type piholeBlocking struct {
	Blocking string `json:"blocking"` // enabled, disabled, failed, unknown
}

type piholeTopDomains struct {
	Domains []struct {
		Domain string `json:"domain"`
		Count  int    `json:"count"`
	} `json:"domains"`
}

// Stats implements usecase.DNSFilterProvider.
func (p *PiHole) Stats(ctx context.Context, s domain.Settings) (domain.Stats, error) {
	var summary piholeSummary
	if err := p.get(ctx, s, "/api/stats/summary", &summary); err != nil {
		return domain.Stats{}, err
	}
	var blocking piholeBlocking
	if err := p.get(ctx, s, "/api/dns/blocking", &blocking); err != nil {
		return domain.Stats{}, err
	}
	var top piholeTopDomains
	if err := p.get(ctx, s, "/api/stats/top_domains?blocked=true&count="+strconv.Itoa(domain.TopBlockedLimit), &top); err != nil {
		return domain.Stats{}, err
	}
	out := domain.Stats{Queries: summary.Queries.Total, Blocked: summary.Queries.Blocked, ProtectionEnabled: blocking.Blocking == "enabled"}
	for _, d := range top.Domains {
		if len(out.TopBlocked) == domain.TopBlockedLimit {
			break
		}
		out.TopBlocked = append(out.TopBlocked, domain.BlockedDomain{Domain: d.Domain, Count: d.Count})
	}
	return out, nil
}

// get reads one endpoint, renewing the session once on 401.
func (p *PiHole) get(ctx context.Context, s domain.Settings, path string, dst any) error {
	err := p.request(ctx, s, path, dst)
	if !errors.Is(err, errUnauthorized) || s.Password == "" {
		return err
	}
	p.forget(s)
	return p.request(ctx, s, path, dst)
}

func (p *PiHole) request(ctx context.Context, s domain.Settings, path string, dst any) error {
	sid, err := p.session(ctx, s)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if sid != "" {
		req.Header.Set("X-FTL-SID", sid)
	}
	return p.do(req, dst)
}

func connectionKey(s domain.Settings) string { return s.URL + "\x00" + s.Password }

func (p *PiHole) forget(s domain.Settings) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.sessions, connectionKey(s))
}

// session returns the cached session ID or logs in; without a password
// Pi-hole needs no session.
func (p *PiHole) session(ctx context.Context, s domain.Settings) (string, error) {
	if s.Password == "" {
		return "", nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if sid, ok := p.sessions[connectionKey(s)]; ok {
		return sid, nil
	}
	body, err := json.Marshal(map[string]string{"password": s.Password})
	if err != nil {
		return "", fmt.Errorf("encode login: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL+"/api/auth", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	var auth struct {
		Session struct {
			Valid bool   `json:"valid"`
			SID   string `json:"sid"`
		} `json:"session"`
	}
	if err := p.do(req, &auth); err != nil {
		if errors.Is(err, errUnauthorized) {
			return "", errPassword
		}
		return "", err
	}
	if !auth.Session.Valid || auth.Session.SID == "" {
		return "", errPassword
	}
	p.sessions[connectionKey(s)] = auth.Session.SID
	return auth.Session.SID, nil
}

func (p *PiHole) do(req *http.Request, dst any) error {
	res, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusUnauthorized:
		return errUnauthorized
	case res.StatusCode != http.StatusOK:
		return fmt.Errorf("unexpected answer from Pi-hole: %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxResponseBytes)).Decode(dst); err != nil {
		return fmt.Errorf("unexpected response from Pi-hole (is it version 6 or newer?): %w", err)
	}
	return nil
}
