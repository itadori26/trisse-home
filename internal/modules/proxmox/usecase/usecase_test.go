package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/proxmox/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type action struct {
	guest domain.Guest
	act   domain.Action
}

type fakeClient struct {
	err     error
	cluster domain.Cluster
	acts    []action
}

func (f *fakeClient) Cluster(_ context.Context, s domain.Settings) (domain.Cluster, error) {
	if f.err != nil {
		return domain.Cluster{}, f.err
	}
	if s.Secret != "secret" {
		return domain.Cluster{}, errors.New("the API token was rejected")
	}
	return f.cluster, nil
}

func (f *fakeClient) Act(_ context.Context, _ domain.Settings, g domain.Guest, a domain.Action) error {
	f.acts = append(f.acts, action{g, a})
	return nil
}

type audit struct{ events []kernel.AuditEvent }

func (a *audit) Audit(_ context.Context, e kernel.AuditEvent) { a.events = append(a.events, e) }

type clk struct{}

func (clk) Now() time.Time { return time.Unix(0, 0) }

func setup() (Proxmox, *memRepo, *fakeClient, *audit) {
	repo := &memRepo{}
	client := &fakeClient{cluster: domain.Cluster{
		Nodes:  []domain.Node{{Name: "pve", Online: true}},
		Guests: []domain.Guest{{ID: 101, Name: "db", Node: "pve", Kind: domain.KindLXC}, {ID: 100, Name: "ha", Node: "pve", Kind: domain.KindVM, Status: "running"}},
	}}
	a := &audit{}
	uc := NewProxmox(Deps{Repo: repo, Client: client, Clock: clk{}, Policy: domain.ActionPolicy{Allowed: []int{100}}, Audit: a})
	return uc, repo, client, a
}

var input = SettingsInput{URL: "https://pve:8006", TokenID: "monitor@pve!th", Secret: "secret"}

func TestSettingsAndStatus(t *testing.T) {
	ctx := context.Background()
	uc, repo, client, _ := setup()
	if _, err := uc.Status(ctx); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	if err := uc.Refresh(ctx); err != nil {
		t.Fatal("refresh without settings is a no-op")
	}
	bad := input
	bad.Secret = "wrong"
	if _, err := uc.SaveSettings(ctx, bad); !errors.Is(err, kernel.ErrInvalid) || repo.s != nil {
		t.Fatalf("a failing connection test must not save: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, input); err != nil {
		t.Fatal(err)
	}
	keep := input
	keep.Secret = ""
	view, err := uc.SaveSettings(ctx, keep)
	if err != nil || !view.HasSecret || repo.s.Secret != "secret" {
		t.Fatalf("an empty secret keeps the stored one: %+v %v", view, err)
	}
	// A changed URL, token ID or fingerprint needs the secret again (#191).
	for name, change := range map[string]func(*SettingsInput){
		"url":         func(in *SettingsInput) { in.URL = "https://10.0.0.99:8006" },
		"token id":    func(in *SettingsInput) { in.TokenID = "other@pve!th" },
		"fingerprint": func(in *SettingsInput) { in.Fingerprint = strings.Repeat("AB", 32) },
	} {
		changed := keep
		change(&changed)
		var v *kernel.ValidationError
		if _, err := uc.SaveSettings(ctx, changed); !errors.As(err, &v) || v.Field != "secret" {
			t.Fatalf("%s: stored secret reused: %v", name, err)
		}
	}
	if s, _ := uc.Settings(ctx); !s.Configured || s.TokenID != "monitor@pve!th" {
		t.Fatalf("settings: %+v", s)
	}
	st, err := uc.Status(ctx)
	if err != nil || st.Cluster.Guests[0].ID != 100 || len(st.Actions) != 1 || st.Actions[0] != 100 {
		t.Fatalf("status is sorted and lists allowed guests: %+v %v", st, err)
	}
	client.err = errors.New("timeout")
	_ = uc.Refresh(ctx)
	if st, _ := uc.Status(ctx); st.Error == "" || len(st.Cluster.Guests) != 2 {
		t.Fatalf("last known state is kept: %+v", st)
	}
}

func TestAct(t *testing.T) {
	ctx := context.Background()
	uc, _, client, a := setup()
	if err := uc.Act(ctx, 100, "start"); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, input); err != nil {
		t.Fatal(err)
	}
	if err := uc.Act(ctx, 101, "start"); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("not allow-listed: %v", err)
	}
	if err := uc.Act(ctx, 100, "stop"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown action: %v", err)
	}
	if err := uc.Act(ctx, 100, "shutdown"); err != nil {
		t.Fatal(err)
	}
	if len(client.acts) != 1 || client.acts[0].guest.Node != "pve" || client.acts[0].guest.Kind != domain.KindVM || client.acts[0].act != domain.ActionShutdown {
		t.Fatalf("node and kind come from the cluster state: %+v", client.acts)
	}
	if len(a.events) != 1 || a.events[0].Type != "proxmox_shutdown" {
		t.Fatalf("audit: %+v", a.events)
	}
	uc2 := NewProxmox(Deps{Repo: &memRepo{s: &domain.Settings{Secret: "secret"}}, Client: &fakeClient{}, Clock: clk{}, Policy: domain.ActionPolicy{Allowed: []int{7}}})
	if err := uc2.Act(ctx, 7, "start"); !errors.Is(err, domain.ErrGuestNotFound) {
		t.Fatalf("unknown guest: %v", err)
	}
}
