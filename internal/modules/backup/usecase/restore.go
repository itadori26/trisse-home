package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"

	"trisse-home/internal/modules/backup/domain"
	"trisse-home/internal/shared/kernel"
)

// Reauthenticator re-checks the password and 2FA code of the signed-in
// account (output port, implemented by the auth module).
type Reauthenticator interface {
	ConfirmIdentity(ctx context.Context, password, code, clientKey string) error
}

// BackupInspector checks a backup file and reads the live schema (output port).
type BackupInspector interface {
	Inspect(ctx context.Context, path string) (domain.BackupInfo, error)
	Current(ctx context.Context) ([]domain.Migration, error)
}

// DatabaseSwapper stages a file and schedules it to replace the database
// on the next start (output port, implemented by the platform).
type DatabaseSwapper interface {
	Stage(ctx context.Context, r io.Reader) (string, error)
	Discard(path string)
	Schedule(path string) error
}

// RestoreSource selects where the backup comes from.
type RestoreSource string

// Restore sources.
const (
	FromLocal  RestoreSource = "local"
	FromRemote RestoreSource = "remote"
	FromUpload RestoreSource = "upload"
)

// RestoreRequest is the input of Restore.
type RestoreRequest struct {
	Password, Code, ClientKey string
	Source                    RestoreSource
	Name                      string    // local or remote backup name
	Upload                    io.Reader // uploaded file
}

// RestoreDeps bundles the ports of the restore.
type RestoreDeps struct {
	Store     Store
	Remote    RemoteStore
	RemoteCfg RemoteSettingsRepository
	Inspector BackupInspector
	Swapper   DatabaseSwapper
	Reauth    Reauthenticator
	Alerts    kernel.AlertSink
	// Restart shuts the server down gracefully so it starts with the
	// restored database (Docker restarts the container).
	Restart func()
}

// Restorer is the input port of RestoreBackup (#46).
type Restorer interface {
	Restore(ctx context.Context, req RestoreRequest) error
}

type restorer struct{ d RestoreDeps }

// NewRestorer returns the restore use case.
func NewRestorer(d RestoreDeps) Restorer { return &restorer{d: d} }

// Restore implements Restorer: confirm identity, stage, validate, schedule
// the swap and restart. Nothing changes when any step fails.
func (uc *restorer) Restore(ctx context.Context, req RestoreRequest) error {
	if p := kernel.PrincipalFrom(ctx); p.IsGuest() || p.IsToken() {
		return domain.ErrRestoreNotAllowed
	}
	if err := uc.d.Reauth.ConfirmIdentity(ctx, req.Password, req.Code, req.ClientKey); err != nil {
		return err
	}
	src, err := uc.open(ctx, req)
	if err != nil {
		return err
	}
	defer src.Close()
	path, err := uc.d.Swapper.Stage(ctx, src)
	if err != nil {
		return kernel.NewPublicErrorf(kernel.ErrInvalid, "The backup could not be received: {detail}", "detail", err.Error())
	}
	scheduled := false
	defer func() {
		if !scheduled {
			uc.d.Swapper.Discard(path)
		}
	}()
	info, err := uc.d.Inspector.Inspect(ctx, path)
	if err != nil {
		return domain.ErrNotADatabase
	}
	current, err := uc.d.Inspector.Current(ctx)
	if err != nil {
		return fmt.Errorf("read current schema: %w", err)
	}
	if err := domain.ValidateRestore(info, current); err != nil {
		return err
	}
	if err := uc.d.Swapper.Schedule(path); err != nil {
		return fmt.Errorf("schedule restore: %w", err)
	}
	scheduled = true
	uc.d.Alerts.Alert(ctx, kernel.Alert{Kind: "security", Title: "Database restore started",
		Body: "trisse-home restarts with a restored backup. The previous database is kept as trisse-home.db.before-restore.", Urgent: true})
	uc.d.Restart()
	return nil
}

func (uc *restorer) open(ctx context.Context, req RestoreRequest) (io.ReadCloser, error) {
	switch req.Source {
	case FromUpload:
		if req.Upload == nil {
			return nil, kernel.NewValidationError("file", "Choose a backup file.")
		}
		return io.NopCloser(req.Upload), nil
	case FromLocal:
		if !domain.ValidName(req.Name) {
			return nil, domain.ErrBackupNotFound
		}
		rc, _, err := uc.d.Store.Open(req.Name)
		return rc, err
	case FromRemote:
		if !domain.ValidName(req.Name) {
			return nil, domain.ErrBackupNotFound
		}
		s, err := uc.d.RemoteCfg.Load(ctx)
		if err != nil {
			return nil, err
		}
		rc, err := uc.d.Remote.Get(ctx, s, req.Name)
		if err != nil {
			if errors.Is(err, kernel.ErrNotFound) {
				return nil, err
			}
			return nil, kernel.TargetFailure(kernel.ErrUnavailable, err)
		}
		return rc, nil
	}
	return nil, kernel.NewValidationError("source", "The source must be local, remote or upload.")
}
