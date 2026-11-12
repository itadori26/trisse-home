package usecase

import (
	"context"
	"sort"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

type memEvents struct{ list []domain.SecurityEvent }

func (m *memEvents) Add(_ context.Context, e domain.SecurityEvent) error {
	m.list = append(m.list, e)
	return nil
}

func (m *memEvents) List(_ context.Context, limit int) ([]domain.SecurityEvent, error) {
	out := append([]domain.SecurityEvent(nil), m.list...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.After(out[j].At) })
	return out[:min(limit, len(out))], nil
}

func (m *memEvents) KnownDevice(_ context.Context, account, ua string) (bool, error) {
	for _, e := range m.list {
		if e.AccountID == account && e.Type == domain.EventLoginSucceeded && e.Client.UserAgent == ua {
			return true, nil
		}
	}
	return false, nil
}

func (m *memEvents) DeleteBefore(_ context.Context, t time.Time) (int64, error) {
	var keep []domain.SecurityEvent
	for _, e := range m.list {
		if !e.At.Before(t) {
			keep = append(keep, e)
		}
	}
	n := int64(len(m.list) - len(keep))
	m.list = keep
	return n, nil
}

type memAlerts struct{ list []kernel.Alert }

func (m *memAlerts) Alert(_ context.Context, a kernel.Alert) { m.list = append(m.list, a) }

func (m *memEvents) types() []domain.EventType {
	out := make([]domain.EventType, 0, len(m.list))
	for _, e := range m.list {
		out = append(out, e.Type)
	}
	return out
}

func withEvents(e *env) (*memEvents, *memAlerts) {
	events, alerts := &memEvents{}, &memAlerts{}
	e.deps.Events = NewRecorder(events, alerts, &seqIDs{}, e.clock, nil)
	return events, alerts
}

func TestLoginEventsAndNewDeviceAlert(t *testing.T) {
	e := newEnv().withAccount()
	events, alerts := withEvents(e)
	login := NewLogin(e.deps)
	req := LoginRequest{Password: goodPassword, ClientKey: "10.0.0.2", UserAgent: "Firefox"}
	if _, err := login.Execute(ctx, req); err != nil {
		t.Fatal(err)
	}
	if _, err := login.Execute(ctx, req); err != nil {
		t.Fatal(err)
	}
	got := events.types()
	want := []domain.EventType{domain.EventLoginSucceeded, domain.EventNewDevice, domain.EventLoginSucceeded}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("events %v, want %v", got, want)
	}
	if len(alerts.list) != 1 || alerts.list[0].Kind != "security" {
		t.Fatalf("exactly one new-device alert expected: %+v", alerts.list)
	}
}

func TestLockoutIsLoggedAndAlerted(t *testing.T) {
	e := newEnv().withAccount()
	events, alerts := withEvents(e)
	login := NewLogin(e.deps)
	for range 3 {
		_, _ = login.Execute(ctx, LoginRequest{Password: "wrong", ClientKey: "10.0.0.9"})
	}
	locked := 0
	for _, ty := range events.types() {
		if ty == domain.EventLoginLocked {
			locked++
		}
	}
	if locked != 1 || len(alerts.list) != 1 || !alerts.list[0].Urgent {
		t.Fatalf("events %v alerts %+v", events.types(), alerts.list)
	}
	for _, ev := range events.list {
		if ev.Detail != "" && ev.Type == domain.EventLoginFailed {
			t.Fatal("failed logins must not store any input")
		}
	}
}

func TestSecurityLogListPruneAudit(t *testing.T) {
	e := newEnv()
	events, _ := withEvents(e)
	log := NewSecurityLog(e.deps.Events)
	log.Audit(ctx, kernel.AuditEvent{Type: "data_imported", Detail: "3 modules"})
	e.clock.now = e.clock.now.Add(domain.EventRetention + time.Hour)
	log.Audit(ctx, kernel.AuditEvent{Type: "data_imported"})
	list, err := log.List(ctx, 0)
	if err != nil || len(list) != 2 || list[0].Detail != "data_imported: " {
		t.Fatalf("%+v %v", list, err)
	}
	if n, _ := log.Prune(ctx); n != 1 || len(events.list) != 1 {
		t.Fatalf("pruned %d", n)
	}
}
