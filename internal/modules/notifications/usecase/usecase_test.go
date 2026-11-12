package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"trisse-home/internal/modules/notifications/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.DefaultSettings(), nil
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeSender struct {
	mu   sync.Mutex
	sent []domain.Message
	err  error
}

func (f *fakeSender) Send(_ context.Context, _ domain.Settings, m domain.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, m)
	return f.err
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func TestNotifier(t *testing.T) {
	ctx := context.Background()
	sender := &fakeSender{}
	n := NewNotifier(Deps{Repo: &memRepo{}, Sender: sender, Clock: clk{}})
	nn := n.(*notifier)
	down := kernel.Observation{Kind: "service", Key: "1", Label: "NAS", Healthy: false, Detail: "timeout"}
	up := down
	up.Healthy = true

	// Not enabled yet: state is tracked but nothing is sent.
	for _, o := range []kernel.Observation{up, down, down} {
		n.Observe(ctx, o)
	}
	nn.Wait()
	if len(sender.sent) != 0 {
		t.Fatal("sent while disabled")
	}
	if err := n.SendTest(ctx); err == nil {
		t.Fatal("test without topic must fail")
	}
	if _, err := n.SaveSettings(ctx, SettingsInput{Enabled: true, Topic: "home", Token: "secret", Rules: domain.DefaultRules()}); err != nil {
		t.Fatal(err)
	}
	n.Observe(ctx, up)
	n.Observe(ctx, up)
	nn.Wait()
	if len(sender.sent) != 1 || sender.sent[0].Title != "NAS is OK again" {
		t.Fatalf("recovery not sent once: %+v", sender.sent)
	}
	view, _ := n.SaveSettings(ctx, SettingsInput{Enabled: true, Topic: "home", Rules: domain.DefaultRules()})
	if !view.HasToken {
		t.Fatal("empty token must keep the stored token")
	}
	sender.err = errors.New("connection refused")
	if err := n.SendTest(ctx); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("test error not reported: %v", err)
	}
	if v, _ := n.Settings(ctx); v.LastError == "" {
		t.Fatal("last delivery error not recorded")
	}
}

func TestAlert(t *testing.T) {
	ctx := context.Background()
	sender := &fakeSender{}
	n := NewNotifier(Deps{Repo: &memRepo{}, Sender: sender, Clock: clk{}})
	nn := n.(*notifier)
	alert := kernel.Alert{Kind: "security", Title: "New sign-in", Urgent: true}
	n.Alert(ctx, alert)
	nn.Wait()
	if len(sender.sent) != 0 {
		t.Fatal("sent while disabled")
	}
	rules := domain.DefaultRules()
	_, _ = n.SaveSettings(ctx, SettingsInput{Enabled: true, Topic: "home", Rules: rules})
	n.Alert(ctx, alert)
	nn.Wait()
	if len(sender.sent) != 1 || sender.sent[0].Severity != domain.SeverityCritical || sender.sent[0].Priority != 5 {
		t.Fatalf("%+v", sender.sent)
	}
	rules.Security = false
	_, _ = n.SaveSettings(ctx, SettingsInput{Enabled: true, Topic: "home", Rules: rules})
	n.Alert(ctx, alert)
	nn.Wait()
	if len(sender.sent) != 1 {
		t.Fatal("security rule off must suppress the alert")
	}
}

type memChannels struct{ list []domain.Channel }

func (m *memChannels) List(context.Context) ([]domain.Channel, error) { return m.list, nil }
func (m *memChannels) Save(_ context.Context, c domain.Channel) error {
	for i := range m.list {
		if m.list[i].ID == c.ID {
			m.list[i] = c
			return nil
		}
	}
	m.list = append(m.list, c)
	return nil
}
func (m *memChannels) Delete(context.Context, string) error { return nil }

type memMaintenance struct{ list []domain.MaintenanceWindow }

func (m *memMaintenance) ListEndingAfter(_ context.Context, t time.Time) ([]domain.MaintenanceWindow, error) {
	var out []domain.MaintenanceWindow
	for _, w := range m.list {
		if w.End.After(t) {
			out = append(out, w)
		}
	}
	return out, nil
}
func (m *memMaintenance) Save(_ context.Context, w domain.MaintenanceWindow) error {
	for i := range m.list {
		if m.list[i].ID == w.ID {
			m.list[i] = w
			return nil
		}
	}
	m.list = append(m.list, w)
	return nil
}
func (m *memMaintenance) Delete(context.Context, string) error { return nil }

type memPending struct{ list []domain.Message }

func (m *memPending) Add(_ context.Context, msg domain.Message, _ time.Time) error {
	m.list = append(m.list, msg)
	return nil
}
func (m *memPending) Take(context.Context) ([]domain.Message, error) {
	out := m.list
	m.list = nil
	return out, nil
}

type recordingPublisher struct {
	mu   sync.Mutex
	sent []domain.Message
}

func (p *recordingPublisher) Publish(_ context.Context, _ domain.Channel, m domain.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sent = append(p.sent, m)
	return nil
}

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

type seq struct{ n int }

func (s *seq) NewID() string { s.n++; return fmt.Sprint(s.n) }

func TestQuietHoursDigestMaintenanceAndRouting(t *testing.T) {
	ctx := context.Background()
	clock := &movingClock{t: time.Date(2026, 11, 9, 23, 0, 0, 0, time.UTC)} // Monday 23:00
	sender, hook := &fakeSender{}, &recordingPublisher{}
	pending, maint := &memPending{}, &memMaintenance{}
	n := NewNotifier(Deps{Repo: &memRepo{}, Sender: sender, Clock: clock, IDs: &seq{}, Channels: &memChannels{},
		Maintenance: maint, Pending: pending, Publishers: map[domain.ChannelKind]Publisher{domain.ChannelWebhook: hook}})
	nn := n.(*notifier)
	policy := domain.Policy{Quiet: domain.QuietHours{Enabled: true, Start: "22:00", End: "07:00"}}
	if _, err := n.SaveSettings(ctx, SettingsInput{Enabled: true, Topic: "home", Rules: domain.DefaultRules(), Policy: policy}); err != nil {
		t.Fatal(err)
	}
	if _, err := n.SaveChannel(ctx, ChannelInput{Channel: domain.Channel{Kind: domain.ChannelWebhook, Name: "Hook", URL: "https://example.org",
		Enabled: true, MinSeverity: domain.SeverityCritical}}); err != nil {
		t.Fatal(err)
	}

	// Night: a warning is deferred, a critical alert goes out immediately
	// (to ntfy and to the webhook that only wants critical ones).
	n.Alert(ctx, kernel.Alert{Kind: "security", Title: "New sign-in"})
	n.Alert(ctx, kernel.Alert{Kind: "security", Title: "Lockout", Urgent: true})
	nn.Wait()
	if len(pending.list) != 1 || len(sender.sent) != 1 || len(hook.sent) != 1 || hook.sent[0].Title != "Lockout" {
		t.Fatalf("pending %d, ntfy %d, webhook %d", len(pending.list), len(sender.sent), len(hook.sent))
	}

	// Still quiet: the digest waits; in the morning it is sent once.
	_ = n.FlushDigest(ctx)
	if len(sender.sent) != 1 {
		t.Fatal("digest sent during quiet hours")
	}
	clock.t = clock.t.Add(9 * time.Hour)
	_ = n.FlushDigest(ctx)
	_ = n.FlushDigest(ctx)
	if len(sender.sent) != 2 || sender.sent[1].Type != "digest" || len(hook.sent) != 1 {
		t.Fatalf("digest: ntfy %+v webhook %d", sender.sent, len(hook.sent))
	}

	// A maintenance window mutes the scoped service.
	if _, err := n.StartMaintenance(ctx, []string{"service:NAS"}, clock.t, clock.t.Add(time.Hour), "update"); err != nil {
		t.Fatal(err)
	}
	nas := kernel.Observation{Kind: "service", Key: "1", Label: "NAS", Healthy: true}
	n.Observe(ctx, nas)
	nas.Healthy = false
	n.Observe(ctx, nas)
	n.Observe(ctx, nas)
	nn.Wait()
	if len(sender.sent) != 2 || !n.UnderMaintenance(ctx, "service", "1", "NAS", clock.t) {
		t.Fatal("maintenance did not mute the service")
	}
	windows, _ := n.Maintenance(ctx)
	if err := n.EndMaintenance(ctx, windows[0].ID); err != nil || n.UnderMaintenance(ctx, "service", "1", "NAS", clock.t) {
		t.Fatalf("end maintenance: %v", err)
	}
}
