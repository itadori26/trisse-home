package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/device/domain"
)

// DisplayDetector tells whether a screen is connected (output port).
type DisplayDetector interface {
	Display(ctx context.Context) domain.Display
}

// ModeStore is the hand-over to the root side (output port): the app writes
// the request and the kiosk address, the system writes the progress.
type ModeStore interface {
	Requested(ctx context.Context) (domain.Mode, error)
	Request(ctx context.Context, m domain.Mode) error
	Progress(ctx context.Context) (domain.Progress, error)
	// SetKioskURL stores the address the display opens ("" removes it).
	SetKioskURL(ctx context.Context, url string) error
}

// KioskIssuer creates the loopback-only kiosk link of this device (output
// port, implemented in the composition root with the auth module).
type KioskIssuer interface {
	Link(ctx context.Context) (token string, err error)
	Remove(ctx context.Context) error
}

// Device is the input port.
type Device interface {
	State(ctx context.Context) (domain.State, error)
	// ChangeMode requests a display mode; the system applies it.
	ChangeMode(ctx context.Context, mode string) (domain.State, error)
}

// Deps bundles the ports.
type Deps struct {
	// Appliance is false outside the appliance image: everything is refused.
	Appliance bool
	Detector  DisplayDetector
	Store     ModeStore
	Kiosk     KioskIssuer
}

type device struct{ d Deps }

// New returns the use cases.
func New(d Deps) Device { return &device{d: d} }

// kioskAddress is what the display opens; the token stays in the fragment.
const kioskAddress = "http://localhost/kiosk#"

func (uc *device) State(ctx context.Context) (domain.State, error) {
	if !uc.d.Appliance {
		return domain.State{}, domain.ErrNotAppliance
	}
	requested, err := uc.d.Store.Requested(ctx)
	if err != nil {
		return domain.State{}, fmt.Errorf("read request: %w", err)
	}
	progress, err := uc.d.Store.Progress(ctx)
	if err != nil {
		return domain.State{}, fmt.Errorf("read progress: %w", err)
	}
	return domain.State{Requested: requested, Progress: progress, Display: uc.d.Detector.Display(ctx)}, nil
}

func (uc *device) ChangeMode(ctx context.Context, raw string) (domain.State, error) {
	s, err := uc.State(ctx)
	if err != nil {
		return domain.State{}, err
	}
	mode, err := domain.ParseMode(raw)
	if err != nil {
		return domain.State{}, err
	}
	if err := s.CanChange(); err != nil {
		return domain.State{}, err
	}
	if mode == domain.ModeKiosk {
		// A fresh link per change: an older one stops working.
		token, err := uc.d.Kiosk.Link(ctx)
		if err != nil {
			return domain.State{}, fmt.Errorf("kiosk link: %w", err)
		}
		if err := uc.d.Store.SetKioskURL(ctx, kioskAddress+token); err != nil {
			return domain.State{}, fmt.Errorf("store kiosk address: %w", err)
		}
	}
	if err := uc.d.Store.Request(ctx, mode); err != nil {
		return domain.State{}, fmt.Errorf("request mode: %w", err)
	}
	if mode == domain.ModeHeadless {
		if err := uc.d.Store.SetKioskURL(ctx, ""); err != nil {
			return domain.State{}, fmt.Errorf("remove kiosk address: %w", err)
		}
		if err := uc.d.Kiosk.Remove(ctx); err != nil {
			return domain.State{}, fmt.Errorf("remove kiosk: %w", err)
		}
	}
	return uc.State(ctx)
}
