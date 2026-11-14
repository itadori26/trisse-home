package gateway

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"trisse-home/internal/modules/device/domain"
)

func TestSysfs(t *testing.T) {
	ctx := context.Background()
	if got := NewSysfs(fstest.MapFS{}).Display(ctx); got != domain.DisplayUnknown {
		t.Fatal(got)
	}
	off := fstest.MapFS{"card0-HDMI-A-1/status": {Data: []byte("disconnected\n")}}
	if got := NewSysfs(off).Display(ctx); got != domain.DisplayDisconnected {
		t.Fatal(got)
	}
	off["card0-DSI-1/status"] = &fstest.MapFile{Data: []byte("connected\n")}
	if got := NewSysfs(off).Display(ctx); got != domain.DisplayConnected {
		t.Fatal(got)
	}
}

func TestFiles(t *testing.T) {
	ctx := context.Background()
	dir := filepath.Join(t.TempDir(), "device")
	f := NewFiles(dir)
	if m, err := f.Requested(ctx); err != nil || m != domain.ModeUnset {
		t.Fatal(m, err)
	}
	if err := f.Request(ctx, domain.ModeKiosk); err != nil {
		t.Fatal(err)
	}
	if m, _ := f.Requested(ctx); m != domain.ModeKiosk {
		t.Fatal(m)
	}
	_ = f.SetKioskURL(ctx, "http://localhost/kiosk#tok")
	info, err := os.Stat(filepath.Join(dir, "kiosk-url"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("the kiosk address holds a credential: %v %v", info.Mode(), err)
	}
	_ = f.SetKioskURL(ctx, "")
	if _, err := os.Stat(filepath.Join(dir, "kiosk-url")); !os.IsNotExist(err) {
		t.Fatal("removed")
	}
	if p, _ := f.Progress(ctx); p.Phase != "" {
		t.Fatal(p)
	}
	_ = os.WriteFile(filepath.Join(dir, "display-mode.status"), []byte(`{"mode":"kiosk","phase":"failed","error":"no internet"}`), 0o644)
	if p, _ := f.Progress(ctx); p.Phase != domain.PhaseFailed || p.Error != "no internet" || p.Mode != domain.ModeKiosk {
		t.Fatal(p)
	}
	_ = os.WriteFile(filepath.Join(dir, "display-mode.status"), []byte(`garbage`), 0o644)
	if p, _ := f.Progress(ctx); p.Phase != domain.PhaseFailed {
		t.Fatal(p)
	}
	_ = os.WriteFile(filepath.Join(dir, "display-mode"), []byte("rm -rf /\n"), 0o644)
	if m, _ := f.Requested(ctx); m != domain.ModeUnset {
		t.Fatal("a broken request counts as not chosen")
	}
}
