package kernel

import "context"

// Alert is a one-off message for the user, e.g. a sign-in from a new device.
// Unlike an Observation it has no state: every alert is delivered once.
type Alert struct {
	// Kind selects the notification rule: "security".
	Kind  string
	Title string
	Body  string
	// Urgent alerts are delivered with high priority.
	Urgent bool
}

// AlertSink receives alerts (output port shared by modules).
type AlertSink interface {
	Alert(ctx context.Context, a Alert)
}

// DiscardAlerts is an AlertSink that ignores everything.
type DiscardAlerts struct{}

// Alert implements AlertSink.
func (DiscardAlerts) Alert(context.Context, Alert) {}

// AuditEvent is a security-relevant action performed in another module,
// e.g. restoring data, for the security event log of the auth module.
type AuditEvent struct {
	// Type is a stable identifier such as "data_imported".
	Type   string
	Detail string
}

// AuditSink records audit events (output port shared by modules).
type AuditSink interface {
	Audit(ctx context.Context, e AuditEvent)
}

// DiscardAudit is an AuditSink that ignores everything.
type DiscardAudit struct{}

// Audit implements AuditSink.
func (DiscardAudit) Audit(context.Context, AuditEvent) {}
