// Package gateway reads the display state from sysfs and exchanges the
// mode with the root side through files in the state directory.
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"trisse-home/internal/modules/device/domain"
)

// Sysfs implements usecase.DisplayDetector on /sys/class/drm.
type Sysfs struct{ fsys fs.FS }

// NewSysfs returns the detector; fsys is /sys/class/drm.
func NewSysfs(fsys fs.FS) *Sysfs { return &Sysfs{fsys: fsys} }

// Display implements usecase.DisplayDetector: connected if any connector
// says so, unknown without connectors (e.g. in a container).
func (s *Sysfs) Display(context.Context) domain.Display {
	matches, err := fs.Glob(s.fsys, "*/status")
	if err != nil || len(matches) == 0 {
		return domain.DisplayUnknown
	}
	out := domain.DisplayDisconnected
	for _, m := range matches {
		data, err := fs.ReadFile(s.fsys, m)
		if err == nil && strings.TrimSpace(string(data)) == "connected" {
			return domain.DisplayConnected
		}
	}
	return out
}

// Files names in the state directory; the root script reads the first two
// and writes the status (see deploy/raspberrypi/rootfs/usr/lib/trisse-home).
const (
	requestFile = "display-mode"
	urlFile     = "kiosk-url"
	statusFile  = "display-mode.status"
)

// Files implements usecase.ModeStore on a directory.
type Files struct{ dir string }

// NewFiles returns the store; dir is created on first write.
func NewFiles(dir string) *Files { return &Files{dir: dir} }

func (f *Files) read(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(f.dir, name)) //nolint:gosec // name is one of the constant file names above, dir comes from the config
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return strings.TrimSpace(string(data)), err
}

// write replaces a file atomically, readable by the owner only.
func (f *Files) write(name, content string) error {
	if err := os.MkdirAll(f.dir, 0o750); err != nil {
		return fmt.Errorf("create %s: %w", f.dir, err)
	}
	tmp, err := os.CreateTemp(f.dir, "."+name+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(f.dir, name))
}

// Requested implements usecase.ModeStore.
func (f *Files) Requested(context.Context) (domain.Mode, error) {
	s, err := f.read(requestFile)
	if err != nil || s == "" {
		return domain.ModeUnset, err
	}
	m, err := domain.ParseMode(s)
	if err != nil {
		return domain.ModeUnset, nil //nolint:nilerr // a broken file counts as not chosen
	}
	return m, nil
}

// Request implements usecase.ModeStore; the path unit reacts to the change.
func (f *Files) Request(_ context.Context, m domain.Mode) error {
	return f.write(requestFile, string(m)+"\n")
}

// Progress implements usecase.ModeStore: {"mode":"kiosk","phase":"active","error":""}.
func (f *Files) Progress(context.Context) (domain.Progress, error) {
	s, err := f.read(statusFile)
	if err != nil || s == "" {
		return domain.Progress{}, err
	}
	var p struct {
		Mode  string `json:"mode"`
		Phase string `json:"phase"`
		Error string `json:"error"`
	}
	if json.Unmarshal([]byte(s), &p) != nil {
		return domain.Progress{Phase: domain.PhaseFailed, Error: "unreadable status from the system"}, nil
	}
	return domain.Progress{Mode: domain.Mode(p.Mode), Phase: domain.Phase(p.Phase), Error: p.Error}, nil
}

// SetKioskURL implements usecase.ModeStore.
func (f *Files) SetKioskURL(_ context.Context, url string) error {
	if url == "" {
		err := os.Remove(filepath.Join(f.dir, urlFile))
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	return f.write(urlFile, url+"\n")
}
