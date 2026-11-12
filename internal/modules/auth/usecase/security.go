package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// SecurityEventRepository persists the security log (output port).
type SecurityEventRepository interface {
	Add(ctx context.Context, e domain.SecurityEvent) error
	// List returns the newest events first.
	List(ctx context.Context, limit int) ([]domain.SecurityEvent, error)
	// KnownDevice reports whether the account ever signed in successfully
	// with this user agent.
	KnownDevice(ctx context.Context, accountID, userAgent string) (bool, error)
	DeleteBefore(ctx context.Context, t time.Time) (int64, error)
}

// SecurityEventView is one log entry for the settings.
type SecurityEventView struct {
	ID        string
	At        time.Time
	Type      string
	AccountID string
	IP        string
	UserAgent string
	Detail    string
}

// SecurityLog is the input port for the security event log.
type SecurityLog interface {
	List(ctx context.Context, limit int) ([]SecurityEventView, error)
	// Prune deletes events older than the retention period (job).
	Prune(ctx context.Context) (int64, error)
	// Audit records an event reported by another module.
	kernel.AuditSink
}

type clientKey struct{}

// WithClient stores the calling client in ctx so security events of
// authenticated actions know where they came from.
func WithClient(ctx context.Context, c domain.Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

func clientFrom(ctx context.Context) domain.Client {
	c, _ := ctx.Value(clientKey{}).(domain.Client)
	return c
}

// Recorder writes security events and sends alerts. Failures are logged and
// never interrupt the action that is being recorded.
type Recorder struct {
	events SecurityEventRepository
	alerts kernel.AlertSink
	ids    kernel.IDGenerator
	clock  kernel.Clock
	log    *slog.Logger
}

func (r *Recorder) record(ctx context.Context, t domain.EventType, accountID string, client domain.Client, detail string) {
	if r == nil || r.events == nil {
		return
	}
	if client == (domain.Client{}) {
		client = clientFrom(ctx)
	}
	e, err := domain.NewSecurityEvent(r.ids.NewID(), r.clock.Now(), t, accountID, client, detail)
	if err == nil {
		err = r.events.Add(ctx, e)
	}
	if err != nil && r.log != nil {
		r.log.WarnContext(ctx, "could not record security event", "type", t, "err", err)
	}
}

func (r *Recorder) alert(ctx context.Context, title, body string, urgent bool) {
	if r == nil || r.alerts == nil {
		return
	}
	r.alerts.Alert(ctx, kernel.Alert{Kind: "security", Title: title, Body: body, Urgent: urgent})
}

// loginSucceeded records the login and alerts on an unknown device.
func (r *Recorder) loginSucceeded(ctx context.Context, accountID string, client domain.Client, method string) {
	if r == nil || r.events == nil {
		return
	}
	known, err := r.events.KnownDevice(ctx, accountID, client.UserAgent)
	r.record(ctx, domain.EventLoginSucceeded, accountID, client, method)
	if err == nil && !known {
		r.record(ctx, domain.EventNewDevice, accountID, client, "")
		r.alert(ctx, "New sign-in to trisse-home", fmt.Sprintf("A new device signed in from %s: %s", client.IP, client.UserAgent), false)
	}
}

// loginFailed records a failure and, when it caused a lockout, alerts.
func (r *Recorder) loginFailed(ctx context.Context, accountID string, client domain.Client, locked bool) {
	r.record(ctx, domain.EventLoginFailed, accountID, client, "")
	if locked {
		r.record(ctx, domain.EventLoginLocked, accountID, client, "")
		r.alert(ctx, "trisse-home: repeated failed logins", fmt.Sprintf("Logins from %s are temporarily blocked after repeated failures.", client.IP), true)
	}
}

type securityLog struct{ r *Recorder }

// NewSecurityLog returns the security log use cases.
func NewSecurityLog(r *Recorder) SecurityLog { return &securityLog{r: r} }

// NewRecorder returns the shared event recorder; alerts may be nil.
func NewRecorder(events SecurityEventRepository, alerts kernel.AlertSink, ids kernel.IDGenerator, clock kernel.Clock, log *slog.Logger) *Recorder {
	return &Recorder{events: events, alerts: alerts, ids: ids, clock: clock, log: log}
}

// List implements SecurityLog.
func (uc *securityLog) List(ctx context.Context, limit int) ([]SecurityEventView, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	list, err := uc.r.events.List(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list security events: %w", err)
	}
	out := make([]SecurityEventView, 0, len(list))
	for _, e := range list {
		out = append(out, SecurityEventView{ID: e.ID, At: e.At, Type: string(e.Type), AccountID: e.AccountID,
			IP: e.Client.IP, UserAgent: e.Client.UserAgent, Detail: e.Detail})
	}
	return out, nil
}

// Prune implements SecurityLog.
func (uc *securityLog) Prune(ctx context.Context) (int64, error) {
	n, err := uc.r.events.DeleteBefore(ctx, uc.r.clock.Now().Add(-domain.EventRetention))
	if err != nil {
		return 0, fmt.Errorf("prune security events: %w", err)
	}
	return n, nil
}

// Audit implements kernel.AuditSink.
func (uc *securityLog) Audit(ctx context.Context, e kernel.AuditEvent) {
	uc.r.record(ctx, domain.EventExternal, "", domain.Client{}, e.Type+": "+e.Detail)
}
