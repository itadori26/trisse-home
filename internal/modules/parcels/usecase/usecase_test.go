package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	ps map[string]domain.Parcel
	s  domain.Settings
}

func (m *memRepo) List(context.Context) ([]domain.Parcel, error) {
	var out []domain.Parcel
	for _, p := range m.ps {
		out = append(out, p)
	}
	return out, nil
}
func (m *memRepo) Get(_ context.Context, id string) (domain.Parcel, error) {
	p, ok := m.ps[id]
	if !ok {
		return p, domain.ErrParcelNotFound
	}
	return p, nil
}
func (m *memRepo) Save(_ context.Context, p domain.Parcel) error         { m.ps[p.ID] = p; return nil }
func (m *memRepo) Delete(_ context.Context, id string) error             { delete(m.ps, id); return nil }
func (m *memRepo) LoadSettings(context.Context) (domain.Settings, error) { return m.s, nil }
func (m *memRepo) SaveSettings(_ context.Context, s domain.Settings) error {
	m.s = s
	return nil
}

type fakeDHL struct {
	calls int
	code  domain.StatusCode
	err   error
}

func (f *fakeDHL) Track(_ context.Context, s domain.Settings, _ string) (domain.Status, Estimate, error) {
	f.calls++
	if s.DHLAPIKey != "key" {
		return domain.Status{}, Estimate{}, kernel.NewPublicError(kernel.ErrUnauthorized, "DHL rejected the API key.")
	}
	return domain.Status{Code: f.code, Description: "In transit", At: time.Unix(1_800_000_000, 0)}, Estimate{To: 1_800_100_000}, f.err
}

type clk struct{}

func (clk) Now() time.Time { return time.Unix(1_800_000_100, 0) }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return string(rune('a' + i.n)) }

func TestParcels(t *testing.T) {
	ctx := context.Background()
	repo, dhl := &memRepo{ps: map[string]domain.Parcel{}, s: domain.Settings{HideDeliveredDays: 3}}, &fakeDHL{code: domain.StatusTransit}
	uc := NewParcels(Deps{Repo: repo, Providers: map[domain.Carrier]Provider{domain.CarrierDHL: dhl}, Clock: clk{}, IDs: &ids{}})

	p, err := uc.Add(ctx, "JJD0003900079 12345", domain.CarrierDHL, "Shoes")
	if err != nil || p.Error == "" || dhl.calls != 0 {
		t.Fatalf("without key the parcel is stored with a hint: %+v %v", p, err)
	}
	if _, err := uc.SaveSettings(ctx, "key", 3); err != nil {
		t.Fatal(err)
	}
	list, _ := uc.List(ctx)
	if len(list) != 1 || list[0].Status.Code != domain.StatusTransit || list[0].Error != "" || list[0].EstimatedDelivery.IsZero() {
		t.Fatalf("after saving the key: %+v", list)
	}
	if v, _ := uc.Settings(ctx); !v.HasDHLAPIKey {
		t.Fatal("key not stored")
	}
	dhl.code = domain.StatusDelivered
	_ = uc.RefreshAll(ctx)
	calls := dhl.calls
	_ = uc.RefreshAll(ctx)
	if dhl.calls != calls {
		t.Fatal("delivered parcels must not be queried again")
	}
	if err := uc.Remove(ctx, "nope"); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := uc.Add(ctx, "x", domain.CarrierDHL, ""); err == nil {
		t.Fatal("invalid number accepted")
	}
}
