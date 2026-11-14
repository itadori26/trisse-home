package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"trisse-home/internal/modules/device/domain"
)

type memStore struct {
	requested domain.Mode
	progress  domain.Progress
	url       string
}

func (m *memStore) Requested(context.Context) (domain.Mode, error)    { return m.requested, nil }
func (m *memStore) Request(_ context.Context, mode domain.Mode) error { m.requested = mode; return nil }
func (m *memStore) Progress(context.Context) (domain.Progress, error) { return m.progress, nil }
func (m *memStore) SetKioskURL(_ context.Context, url string) error   { m.url = url; return nil }

type fixedDisplay domain.Display

func (f fixedDisplay) Display(context.Context) domain.Display { return domain.Display(f) }

type fakeKiosk struct {
	links   int
	removed bool
}

func (f *fakeKiosk) Link(context.Context) (string, error) {
	f.links++
	return "tok" + string(rune('0'+f.links)), nil
}
func (f *fakeKiosk) Remove(context.Context) error { f.removed = true; return nil }

func TestDevice(t *testing.T) {
	ctx := context.Background()
	if _, err := New(Deps{}).State(ctx); !errors.Is(err, domain.ErrNotAppliance) {
		t.Fatal("not an appliance", err)
	}
	store := &memStore{}
	kiosk := &fakeKiosk{}
	uc := New(Deps{Appliance: true, Detector: fixedDisplay(domain.DisplayConnected), Store: store, Kiosk: kiosk})
	s, err := uc.State(ctx)
	if err != nil || s.Phase() != domain.PhaseIdle || s.Display != domain.DisplayConnected {
		t.Fatalf("%+v %v", s, err)
	}
	s, err = uc.ChangeMode(ctx, "kiosk")
	if err != nil || s.Phase() != domain.PhasePending || store.url != "http://localhost/kiosk#tok1" || kiosk.links != 1 {
		t.Fatalf("%+v %v %q", s, err, store.url)
	}
	store.progress = domain.Progress{Mode: domain.ModeKiosk, Phase: domain.PhaseInstalling}
	if _, err := uc.ChangeMode(ctx, "headless"); !errors.Is(err, domain.ErrBusy) {
		t.Fatal("busy", err)
	}
	store.progress.Phase = domain.PhaseActive
	s, err = uc.ChangeMode(ctx, "headless")
	if err != nil || store.requested != domain.ModeHeadless || store.url != "" || !kiosk.removed || s.Phase() != domain.PhasePending {
		t.Fatalf("%+v %v", s, err)
	}
	if _, err := uc.ChangeMode(ctx, "desktop"); err == nil || !strings.Contains(err.Error(), "headless or kiosk") {
		t.Fatal(err)
	}
}
