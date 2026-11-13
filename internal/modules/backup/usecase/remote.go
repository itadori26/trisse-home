package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/shared/kernel"
)

// RemoteStore talks to the off-site target, e.g. WebDAV (output port).
type RemoteStore interface {
	// Test checks that the target collection is reachable and writable.
	Test(ctx context.Context, s domain.RemoteSettings) error
	Put(ctx context.Context, s domain.RemoteSettings, name string, r io.Reader, size int64) error
	// List returns the backups in the target collection.
	List(ctx context.Context, s domain.RemoteSettings) ([]domain.File, error)
	Delete(ctx context.Context, s domain.RemoteSettings, name string) error
	Get(ctx context.Context, s domain.RemoteSettings, name string) (io.ReadCloser, error)
}

// RemoteSettingsRepository persists the target with an encrypted password
// (output port); Load returns domain.ErrRemoteNotConfigured when unset.
type RemoteSettingsRepository interface {
	Load(ctx context.Context) (domain.RemoteSettings, error)
	Save(ctx context.Context, s domain.RemoteSettings) error
	Delete(ctx context.Context) error
}

// RemoteSettingsView never contains the password.
type RemoteSettingsView struct {
	Configured  bool
	URL         string
	Username    string
	HasPassword bool
	Retention   int
}

// RemoteSettingsInput is the input of SaveSettings and Test; an empty
// password keeps the stored one.
type RemoteSettingsInput struct {
	URL, Username, Password string
	Retention               int
}

// Offsite is the input port of the off-site backups (#46).
type Offsite interface {
	Settings(ctx context.Context) (RemoteSettingsView, error)
	// SaveSettings tests the connection before saving.
	SaveSettings(ctx context.Context, in RemoteSettingsInput) (RemoteSettingsView, error)
	// Test checks the connection without saving.
	Test(ctx context.Context, in RemoteSettingsInput) error
	Disable(ctx context.Context) error
	// List returns the off-site backups, newest first (ListRemoteBackups).
	List(ctx context.Context) ([]FileView, error)
	// Upload copies a local backup and prunes old remote copies
	// (UploadBackup, PruneRemote). It does nothing when not configured.
	Upload(ctx context.Context, name string) error
}

type offsite struct {
	store  Store
	remote RemoteStore
	repo   RemoteSettingsRepository
	alerts kernel.AlertSink
}

// NewOffsite returns the off-site use cases; alerts receive upload failures.
func NewOffsite(store Store, remote RemoteStore, repo RemoteSettingsRepository, alerts kernel.AlertSink) Offsite {
	return &offsite{store: store, remote: remote, repo: repo, alerts: alerts}
}

func remoteView(s domain.RemoteSettings) RemoteSettingsView {
	return RemoteSettingsView{Configured: true, URL: s.URL, Username: s.Username, HasPassword: s.Password != "", Retention: s.Retention}
}

// Settings implements Offsite.
func (uc *offsite) Settings(ctx context.Context) (RemoteSettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrRemoteNotConfigured) {
		return RemoteSettingsView{Retention: domain.DefaultRemoteRetention}, nil
	}
	if err != nil {
		return RemoteSettingsView{}, fmt.Errorf("load off-site settings: %w", err)
	}
	return remoteView(s), nil
}

func (uc *offsite) settingsFrom(ctx context.Context, in RemoteSettingsInput) (domain.RemoteSettings, error) {
	password := in.Password
	if password == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			if kernel.TargetChanged([]string{old.URL, old.Username}, []string{in.URL, in.Username}) {
				return domain.RemoteSettings{}, kernel.SecretRequired("password")
			}
			password = old.Password
		}
	}
	return domain.NewRemoteSettings(in.URL, in.Username, password, in.Retention)
}

// Test implements Offsite.
func (uc *offsite) Test(ctx context.Context, in RemoteSettingsInput) error {
	s, err := uc.settingsFrom(ctx, in)
	if err != nil {
		return err
	}
	if err := uc.remote.Test(ctx, s); err != nil {
		return kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	return nil
}

// SaveSettings implements Offsite.
func (uc *offsite) SaveSettings(ctx context.Context, in RemoteSettingsInput) (RemoteSettingsView, error) {
	s, err := uc.settingsFrom(ctx, in)
	if err != nil {
		return RemoteSettingsView{}, err
	}
	if err := uc.remote.Test(ctx, s); err != nil {
		return RemoteSettingsView{}, kernel.TargetFailure(kernel.ErrInvalid, err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return RemoteSettingsView{}, fmt.Errorf("save off-site settings: %w", err)
	}
	return remoteView(s), nil
}

// Disable implements Offsite.
func (uc *offsite) Disable(ctx context.Context) error { return uc.repo.Delete(ctx) }

// List implements Offsite.
func (uc *offsite) List(ctx context.Context) ([]FileView, error) {
	s, err := uc.repo.Load(ctx)
	if err != nil {
		return nil, err
	}
	files, err := uc.remote.List(ctx, s)
	if err != nil {
		return nil, kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	slices.SortFunc(files, func(a, b domain.File) int { return b.CreatedAt.Compare(a.CreatedAt) })
	out := make([]FileView, 0, len(files))
	for _, f := range files {
		out = append(out, toView(f))
	}
	return out, nil
}

// Upload implements Offsite. Failures are reported as alerts.
func (uc *offsite) Upload(ctx context.Context, name string) error {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrRemoteNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load off-site settings: %w", err)
	}
	if err := uc.upload(ctx, s, name); err != nil {
		uc.alerts.Alert(ctx, kernel.Alert{Kind: "backup", Title: "Off-site backup failed", Body: err.Error(), Urgent: true})
		return err
	}
	return nil
}

func (uc *offsite) upload(ctx context.Context, s domain.RemoteSettings, name string) error {
	rc, f, err := uc.store.Open(name)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer rc.Close()
	if err := uc.remote.Put(ctx, s, name, rc, f.Size); err != nil {
		return fmt.Errorf("upload %s: %w", name, err)
	}
	files, err := uc.remote.List(ctx, s)
	if err != nil {
		return fmt.Errorf("list off-site backups: %w", err)
	}
	for _, old := range s.RemotePolicy().ToPrune(files) {
		if err := uc.remote.Delete(ctx, s, old.Name); err != nil {
			return fmt.Errorf("delete old off-site backup %s: %w", old.Name, err)
		}
	}
	return nil
}
