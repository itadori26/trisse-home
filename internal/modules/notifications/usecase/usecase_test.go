package usecase

import (
	"context"
	"errors"
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
	n := NewNotifier(&memRepo{}, sender, clk{})
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
