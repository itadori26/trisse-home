package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/shared/kernel"
)

// SettingsView never contains the API key.
type SettingsView struct {
	HasDHLAPIKey      bool
	HideDeliveredDays int
}

// Parcels is the input port of the module.
type Parcels interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, dhlKey string, hideDays int) (SettingsView, error)
	List(ctx context.Context) ([]domain.Parcel, error)
	Add(ctx context.Context, trackingNumber string, carrier domain.Carrier, label string) (domain.Parcel, error)
	Remove(ctx context.Context, id string) error
	// RefreshAll updates every undelivered parcel (background job).
	RefreshAll(ctx context.Context) error
	// MailSettings and SaveMailSettings configure the mail discovery (#166).
	MailSettings(ctx context.Context) (MailSettingsView, error)
	SaveMailSettings(ctx context.Context, in MailSettingsInput) (MailSettingsView, error)
	// DiscoverParcels reads new shipping mails and adds their parcels
	// (background job; does nothing while disabled or before the interval).
	DiscoverParcels(ctx context.Context) error
}

// Repository persists parcels and settings (output port).
type Repository interface {
	List(ctx context.Context) ([]domain.Parcel, error)
	Get(ctx context.Context, id string) (domain.Parcel, error)
	Save(ctx context.Context, p domain.Parcel) error
	Delete(ctx context.Context, id string) error
	LoadSettings(ctx context.Context) (domain.Settings, error)
	SaveSettings(ctx context.Context, s domain.Settings) error
}

// Deps bundles the output ports.
type Deps struct {
	Repo      Repository
	Providers map[domain.Carrier]Provider
	Clock     kernel.Clock
	IDs       kernel.IDGenerator
	// Mail and MailStore are the mail discovery (#166).
	Mail      MailSource
	MailStore MailStore
}

// Provider asks one carrier for the latest status (output port; one
// implementation per carrier, chosen by domain.Carrier).
type Provider interface {
	Track(ctx context.Context, s domain.Settings, trackingNumber string) (domain.Status, Estimate, error)
}

// Estimate is the estimated delivery window as Unix seconds (zero if unknown).
type Estimate struct{ From, To int64 }

type parcels struct{ d Deps }

// NewParcels returns the parcel tracking use cases.
func NewParcels(d Deps) Parcels { return &parcels{d: d} }

// Settings implements Parcels.
func (uc *parcels) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.d.Repo.LoadSettings(ctx)
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{HasDHLAPIKey: s.DHLAPIKey != "", HideDeliveredDays: s.HideDeliveredDays}, nil
}

// SaveSettings implements Parcels; an empty key keeps the stored one.
func (uc *parcels) SaveSettings(ctx context.Context, dhlKey string, hideDays int) (SettingsView, error) {
	old, err := uc.d.Repo.LoadSettings(ctx)
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	if dhlKey == "" {
		dhlKey = old.DHLAPIKey
	}
	s, err := domain.NewSettings(dhlKey, hideDays)
	if err != nil {
		return SettingsView{}, err
	}
	if err := uc.d.Repo.SaveSettings(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	if s.DHLAPIKey != old.DHLAPIKey {
		_ = uc.RefreshAll(ctx) // show the status right away; errors are stored per parcel
	}
	return SettingsView{HasDHLAPIKey: s.DHLAPIKey != "", HideDeliveredDays: s.HideDeliveredDays}, nil
}

// List implements Parcels.
func (uc *parcels) List(ctx context.Context) ([]domain.Parcel, error) {
	s, err := uc.d.Repo.LoadSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	all, err := uc.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list parcels: %w", err)
	}
	now := uc.d.Clock.Now()
	out := make([]domain.Parcel, 0, len(all))
	for _, p := range all {
		if p.Visible(now, s.HideDeliveredDays) {
			out = append(out, p)
		}
	}
	domain.SortForDisplay(out)
	return out, nil
}

// Add implements Parcels; the parcel is tracked immediately.
func (uc *parcels) Add(ctx context.Context, trackingNumber string, carrier domain.Carrier, label string) (domain.Parcel, error) {
	return uc.add(ctx, trackingNumber, carrier, label, false)
}

func (uc *parcels) add(ctx context.Context, trackingNumber string, carrier domain.Carrier, label string, discovered bool) (domain.Parcel, error) {
	all, err := uc.d.Repo.List(ctx)
	if err != nil {
		return domain.Parcel{}, fmt.Errorf("list parcels: %w", err)
	}
	if len(all) >= domain.MaxParcels {
		return domain.Parcel{}, domain.ErrTooMany
	}
	p, err := domain.NewParcel(uc.d.IDs.NewID(), trackingNumber, carrier, label, uc.d.Clock.Now())
	if err != nil {
		return domain.Parcel{}, err
	}
	p.Discovered = discovered
	s, err := uc.d.Repo.LoadSettings(ctx)
	if err != nil {
		return domain.Parcel{}, fmt.Errorf("load settings: %w", err)
	}
	p = uc.track(ctx, s, p)
	if err := uc.d.Repo.Save(ctx, p); err != nil {
		return domain.Parcel{}, fmt.Errorf("save parcel: %w", err)
	}
	return p, nil
}

// Remove implements Parcels. A removed discovered parcel is not found
// again in the next run.
func (uc *parcels) Remove(ctx context.Context, id string) error {
	p, err := uc.d.Repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if p.Discovered && uc.d.MailStore != nil {
		if err := uc.d.MailStore.Dismiss(ctx, p.TrackingNumber); err != nil {
			return fmt.Errorf("dismiss number: %w", err)
		}
	}
	return uc.d.Repo.Delete(ctx, id)
}

// RefreshAll implements Parcels.
func (uc *parcels) RefreshAll(ctx context.Context) error {
	s, err := uc.d.Repo.LoadSettings(ctx)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	all, err := uc.d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list parcels: %w", err)
	}
	var failed int
	for _, p := range all {
		if _, trackable := uc.d.Providers[p.Carrier]; p.Delivered() || !trackable {
			continue
		}
		p = uc.track(ctx, s, p)
		if p.Error != "" {
			failed++
		}
		if err := uc.d.Repo.Save(ctx, p); err != nil {
			return fmt.Errorf("save parcel: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d parcels could not be tracked: %w", failed, kernel.ErrUnavailable)
	}
	return nil
}

// track asks the carrier and stores the result or the error on the parcel.
func (uc *parcels) track(ctx context.Context, s domain.Settings, p domain.Parcel) domain.Parcel {
	provider, ok := uc.d.Providers[p.Carrier]
	switch {
	case !ok:
		// No tracking API (yet): the UI links to the carrier's page (#166).
		return p
	case p.Carrier == domain.CarrierDHL && s.DHLAPIKey == "":
		p.Error = domain.ErrNoAPIKey.Error()
		return p
	}
	status, est, err := provider.Track(ctx, s, p.TrackingNumber)
	p.CheckedAt = uc.d.Clock.Now()
	if err != nil {
		var pe *kernel.PublicError
		if errors.As(err, &pe) {
			p.Error = pe.Error()
		} else {
			p.Error = "The carrier could not be reached."
		}
		return p
	}
	p.Status, p.Error = status, ""
	if est.To > 0 {
		p.EstimatedDelivery = unix(est.To)
	}
	return p
}

func unix(sec int64) time.Time { return time.Unix(sec, 0).UTC() }
