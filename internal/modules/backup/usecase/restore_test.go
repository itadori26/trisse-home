package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeReauth struct{ err error }

func (f fakeReauth) ConfirmIdentity(context.Context, string, string, string) error { return f.err }

type fakeInspector struct {
	info    domain.BackupInfo
	err     error
	current []domain.Migration
}

func (f fakeInspector) Inspect(context.Context, string) (domain.BackupInfo, error) {
	return f.info, f.err
}
func (f fakeInspector) Current(context.Context) ([]domain.Migration, error) { return f.current, nil }

type fakeSwapper struct {
	staged, discarded, scheduled string
	content                      string
}

func (f *fakeSwapper) Stage(_ context.Context, r io.Reader) (string, error) {
	b, _ := io.ReadAll(r)
	f.content, f.staged = string(b), "/data/.restore-upload-1"
	return f.staged, nil
}
func (f *fakeSwapper) Discard(p string) { f.discarded = p }
func (f *fakeSwapper) Schedule(p string) error {
	f.scheduled = p
	return nil
}

func TestRestore(t *testing.T) {
	ctx := context.Background()
	migrations := []domain.Migration{{Module: "auth", Version: "0001"}}
	good := domain.BackupInfo{Integrity: "ok", Migrations: migrations}
	name := "trisse-home-20261111-030000.db"
	store := &memStore{files: map[string]domain.File{name: {Name: name, CreatedAt: time.Now()}}}

	run := func(reauth error, info domain.BackupInfo, req RestoreRequest, c context.Context) (*fakeSwapper, *alertRec, bool, error) {
		sw, alerts, restarted := &fakeSwapper{}, &alertRec{}, false
		uc := NewRestorer(RestoreDeps{Store: store, Inspector: fakeInspector{info: info, current: migrations}, Swapper: sw,
			Reauth: fakeReauth{reauth}, Alerts: alerts, Restart: func() { restarted = true }})
		return sw, alerts, restarted, uc.Restore(c, req)
	}

	sw, alerts, restarted, err := run(nil, good, RestoreRequest{Source: FromUpload, Upload: strings.NewReader("sqlite")}, ctx)
	if err != nil || !restarted || sw.scheduled != sw.staged || sw.discarded != "" || sw.content != "sqlite" || len(alerts.got) != 1 {
		t.Fatalf("upload: %v %v %+v", err, restarted, sw)
	}
	if sw, _, _, err := run(nil, good, RestoreRequest{Source: FromLocal, Name: name}, ctx); err != nil || sw.content != "data" {
		t.Fatalf("local: %v %+v", err, sw)
	}

	// Wrong password: nothing is read or staged.
	sw, _, restarted, err = run(domain.ErrBackupNotFound, good, RestoreRequest{Source: FromUpload, Upload: strings.NewReader("x")}, ctx)
	if err == nil || restarted || sw.staged != "" {
		t.Fatalf("reauth: %v %+v", err, sw)
	}
	// A newer backup is rejected and the staged file removed.
	newer := domain.BackupInfo{Integrity: "ok", Migrations: append([]domain.Migration{{Module: "auth", Version: "0002"}}, migrations...)}
	sw, _, restarted, err = run(nil, newer, RestoreRequest{Source: FromUpload, Upload: strings.NewReader("x")}, ctx)
	if !errors.Is(err, domain.ErrTooNew) || restarted || sw.scheduled != "" || sw.discarded != sw.staged {
		t.Fatalf("too new: %v %+v", err, sw)
	}
	if _, _, _, err := run(nil, good, RestoreRequest{Source: FromLocal, Name: "../../etc/passwd"}, ctx); !errors.Is(err, domain.ErrBackupNotFound) {
		t.Fatalf("traversal: %v", err)
	}
	guest := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "g", Role: "guest"})
	if _, _, _, err := run(nil, good, RestoreRequest{Source: FromUpload, Upload: strings.NewReader("x")}, guest); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("guest: %v", err)
	}
	token := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "owner", Role: "token"})
	if _, _, _, err := run(nil, good, RestoreRequest{Source: FromUpload, Upload: strings.NewReader("x")}, token); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("token: %v", err)
	}
}
