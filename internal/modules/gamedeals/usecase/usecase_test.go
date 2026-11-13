package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/shared/kernel"
)

var now = time.Date(2026, 11, 12, 16, 0, 0, 0, time.UTC)

type memRepo struct{ s domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error)   { return m.s, nil }
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = s; return nil }

type fakeSteam struct {
	err   error
	calls int
}

func (f *fakeSteam) Wishlist(context.Context, string, string) ([]domain.Deal, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return []domain.Deal{{Name: "Big", DiscountPct: 75}, {Name: "Small", DiscountPct: 10}}, nil
}

type fakeEpic struct{ err error }

func (f *fakeEpic) FreeGames(context.Context, string) ([]domain.FreeGame, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []domain.FreeGame{
		{Title: "Now", Start: now.Add(-24 * time.Hour), End: now.Add(24 * time.Hour)},
		{Title: "Next", Start: now.Add(6 * 24 * time.Hour), End: now.Add(13 * 24 * time.Hour)},
		{Title: "Over", Start: now.Add(-9 * 24 * time.Hour), End: now.Add(-2 * 24 * time.Hour)},
	}, nil
}

type clk struct{}

func (clk) Now() time.Time { return now }

func TestGameDeals(t *testing.T) {
	ctx := context.Background()
	repo, steam, epic := &memRepo{s: domain.DefaultSettings()}, &fakeSteam{}, &fakeEpic{}
	uc := NewGameDeals(Deps{Repo: repo, Wishlist: steam, Free: epic, Clock: clk{}})

	if err := uc.Refresh(ctx); err != nil || steam.calls != 0 {
		t.Fatal("nothing is requested while every source is off")
	}
	steam.err = errors.New("403 Forbidden")
	if _, err := uc.SaveSettings(ctx, domain.Settings{SteamID: "76561197960287930", Country: "DE", MinDiscount: 20}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("a private wishlist is reported: %v", err)
	}
	steam.err = nil
	if _, err := uc.SaveSettings(ctx, domain.Settings{SteamID: "76561197960287930", Country: "de", MinDiscount: 20, Epic: true}); err != nil {
		t.Fatal(err)
	}
	o, err := uc.Overview(ctx)
	if err != nil || len(o.Deals) != 1 || o.Deals[0].Name != "Big" {
		t.Fatalf("deals above the minimum: %+v %v", o, err)
	}
	if len(o.FreeNow) != 1 || o.FreeNow[0].Title != "Now" || len(o.FreeNext) != 1 || o.FreeNext[0].Title != "Next" {
		t.Fatalf("free games now and next week, expired ones dropped: %+v", o)
	}

	epic.err = errors.New("shape changed")
	if err := uc.Refresh(ctx); err == nil {
		t.Fatal("source errors are reported to the log")
	}
	o, _ = uc.Overview(ctx)
	if len(o.Deals) != 1 || len(o.FreeNow) != 1 || len(o.Errors) != 1 || o.Errors[0] != "Epic Games Store" {
		t.Fatalf("a failing source keeps its last data and does not hide the other: %+v", o)
	}
}
