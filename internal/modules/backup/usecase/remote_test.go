package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/shared/kernel"
)

type memRemote struct {
	files   map[string][]byte
	created map[string]time.Time
	testErr error
	putErr  error
}

func (m *memRemote) Test(context.Context, domain.RemoteSettings) error { return m.testErr }
func (m *memRemote) Put(_ context.Context, _ domain.RemoteSettings, name string, r io.Reader, _ int64) error {
	if m.putErr != nil {
		return m.putErr
	}
	b, _ := io.ReadAll(r)
	m.files[name] = b
	return nil
}
func (m *memRemote) List(context.Context, domain.RemoteSettings) ([]domain.File, error) {
	var out []domain.File
	for n := range m.files {
		out = append(out, domain.File{Name: n, CreatedAt: m.created[n]})
	}
	return out, nil
}
func (m *memRemote) Delete(_ context.Context, _ domain.RemoteSettings, name string) error {
	delete(m.files, name)
	return nil
}
func (m *memRemote) Get(_ context.Context, _ domain.RemoteSettings, name string) (io.ReadCloser, error) {
	b, ok := m.files[name]
	if !ok {
		return nil, domain.ErrBackupNotFound
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

type memRemoteSettings struct{ s *domain.RemoteSettings }

func (m *memRemoteSettings) Load(context.Context) (domain.RemoteSettings, error) {
	if m.s == nil {
		return domain.RemoteSettings{}, domain.ErrRemoteNotConfigured
	}
	return *m.s, nil
}
func (m *memRemoteSettings) Save(_ context.Context, s domain.RemoteSettings) error {
	m.s = &s
	return nil
}
func (m *memRemoteSettings) Delete(context.Context) error { m.s = nil; return nil }

type alertRec struct{ got []kernel.Alert }

func (a *alertRec) Alert(_ context.Context, al kernel.Alert) { a.got = append(a.got, al) }

func TestOffsite(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 11, 3, 0, 0, 0, time.UTC)}
	store := &memStore{files: map[string]domain.File{}}
	remote := &memRemote{files: map[string][]byte{}, created: map[string]time.Time{}}
	repo := &memRemoteSettings{}
	alerts := &alertRec{}
	off := NewOffsite(store, remote, repo, alerts)
	policy, _ := domain.NewPolicy(time.Hour, 10)
	uc := NewBackups(policy, &fakeSnap{store: store, clock: c}, store, c, WithAfterBackup(func(ctx context.Context, f FileView) {
		remote.created[f.Name] = f.CreatedAt
		_ = off.Upload(ctx, f.Name)
	}))

	// Not configured: backups work, nothing is uploaded.
	if _, err := uc.BackupNow(ctx); err != nil || len(remote.files) != 0 {
		t.Fatalf("unconfigured: %v %d", err, len(remote.files))
	}
	if v, _ := off.Settings(ctx); v.Configured || v.Retention != domain.DefaultRemoteRetention {
		t.Fatalf("defaults: %+v", v)
	}
	remote.testErr = errors.New("401")
	if _, err := off.SaveSettings(ctx, RemoteSettingsInput{URL: "https://cloud/dav/", Username: "me", Password: "pw", Retention: 2}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("failing test must not save: %v", err)
	}
	remote.testErr = nil
	if _, err := off.SaveSettings(ctx, RemoteSettingsInput{URL: "https://cloud/dav/", Username: "me", Password: "pw", Retention: 2}); err != nil {
		t.Fatal(err)
	}
	if err := off.Test(ctx, RemoteSettingsInput{URL: "https://cloud/dav/", Username: "me", Retention: 2}); err != nil {
		t.Fatalf("test with stored password: %v", err)
	}
	for range 3 {
		c.t = c.t.Add(time.Hour)
		if _, err := uc.BackupNow(ctx); err != nil {
			t.Fatal(err)
		}
	}
	// Local retention 10, remote retention 2.
	if len(store.files) != 4 || len(remote.files) != 2 {
		t.Fatalf("local %d, remote %d", len(store.files), len(remote.files))
	}
	list, _ := off.List(ctx)
	if len(list) != 2 || !list[0].CreatedAt.After(list[1].CreatedAt) {
		t.Fatalf("list: %+v", list)
	}
	remote.putErr = errors.New("507 insufficient storage")
	c.t = c.t.Add(time.Hour)
	_, _ = uc.BackupNow(ctx)
	if len(alerts.got) != 1 || alerts.got[0].Kind != "backup" || !alerts.got[0].Urgent {
		t.Fatalf("alerts: %+v", alerts.got)
	}
	_ = off.Disable(ctx)
	if v, _ := off.Settings(ctx); v.Configured {
		t.Fatal("disable")
	}
}
