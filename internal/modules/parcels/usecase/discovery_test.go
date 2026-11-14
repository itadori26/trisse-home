package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/shared/kernel"
)

type fakeMail struct {
	mails   []domain.Mail
	err     error
	fetches int
	since   time.Time
	tested  domain.MailSettings
}

func (f *fakeMail) Fetch(_ context.Context, _ domain.MailSettings, st MailState, since time.Time) ([]domain.Mail, MailState, error) {
	f.fetches++
	f.since = since
	if f.err != nil {
		return nil, st, f.err
	}
	return f.mails, MailState{UIDValidity: 1, LastUID: st.LastUID + uint32(len(f.mails))}, nil
}

func (f *fakeMail) Test(_ context.Context, s domain.MailSettings) error {
	f.tested = s
	return f.err
}

type memMail struct {
	s         domain.MailSettings
	st        MailState
	dismissed map[string]bool
}

func (m *memMail) LoadMail(context.Context) (domain.MailSettings, MailState, error) {
	return m.s, m.st, nil
}
func (m *memMail) SaveMail(_ context.Context, s domain.MailSettings, st MailState) error {
	m.s, m.st = s, st
	return nil
}
func (m *memMail) Dismissed(_ context.Context, n string) (bool, error) { return m.dismissed[n], nil }
func (m *memMail) Dismiss(_ context.Context, n string) error {
	m.dismissed[n] = true
	return nil
}

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

func TestDiscoverParcels(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{ps: map[string]domain.Parcel{}, s: domain.Settings{DHLAPIKey: "key"}}
	mail := &fakeMail{mails: []domain.Mail{
		{From: "DHL <noreply@dhl.de>", Subject: "Ihre Sendung von Zalando", Text: "Sendungsnummer: 00340434161094042557"},
		{From: "Shop <a@shop.example>", Subject: "Versand", Text: "Mit Hermes, Sendungsnummer 01234567890123"},
		{From: "DHL <noreply@dhl.de>", Subject: "Erinnerung", Text: "Sendungsnummer: 00340434161094042557"},
	}}
	store := &memMail{dismissed: map[string]bool{}, s: domain.MailSettings{Interval: 15 * time.Minute}}
	clock := &movingClock{t: time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)}
	dhl := &fakeDHL{code: domain.StatusTransit}
	uc := NewParcels(Deps{Repo: repo, Providers: map[domain.Carrier]Provider{domain.CarrierDHL: dhl}, Clock: clock, IDs: &ids{}, Mail: mail, MailStore: store})

	// Disabled: nothing happens.
	if err := uc.DiscoverParcels(ctx); err != nil || mail.fetches != 0 {
		t.Fatal("disabled must not connect", err)
	}
	// Enabling tests the connection; an empty password needs the old one for the same server.
	if _, err := uc.SaveMailSettings(ctx, MailSettingsInput{Enabled: true, Host: "imap.example.com", Username: "me", Folder: "Pakete"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("password required", err)
	}
	v, err := uc.SaveMailSettings(ctx, MailSettingsInput{Enabled: true, Host: "imap.example.com", Username: "me", Password: "pw", Folder: "Pakete", IntervalMinutes: 15})
	if err != nil || !v.HasPassword || mail.tested.Host != "imap.example.com" {
		t.Fatalf("%+v %v", v, err)
	}
	if err := uc.DiscoverParcels(ctx); err != nil {
		t.Fatal(err)
	}
	if !mail.since.Equal(clock.t.AddDate(0, 0, -14)) || len(repo.ps) != 2 || store.st.Found != 2 || store.st.LastUID != 3 {
		t.Fatalf("since %v, %d parcels, state %+v", mail.since, len(repo.ps), store.st)
	}
	var hermes, dhlParcel domain.Parcel
	for _, p := range repo.ps {
		if p.Carrier == domain.CarrierHermes {
			hermes = p
		} else {
			dhlParcel = p
		}
	}
	if !hermes.Discovered || hermes.Error != "" || hermes.Label != "Shop" || dhlParcel.Status.Code != domain.StatusTransit || dhlParcel.Label != "Zalando" {
		t.Fatalf("hermes %+v, dhl %+v", hermes, dhlParcel)
	}
	// Before the interval nothing is fetched.
	clock.t = clock.t.Add(10 * time.Minute)
	_ = uc.DiscoverParcels(ctx)
	if mail.fetches != 1 {
		t.Fatal("interval")
	}
	// A removed discovered parcel is not added again.
	if err := uc.Remove(ctx, hermes.ID); err != nil || !store.dismissed["01234567890123"] {
		t.Fatal("dismiss", err)
	}
	clock.t = clock.t.Add(10 * time.Minute)
	if err := uc.DiscoverParcels(ctx); err != nil || len(repo.ps) != 1 || store.st.Found != 2 {
		t.Fatalf("re-added? %d parcels, %+v, %v", len(repo.ps), store.st, err)
	}
	// RefreshAll skips carriers without a provider and does not count them as failures.
	repo.ps["z"] = domain.Parcel{ID: "z", TrackingNumber: "12345678901", Carrier: domain.CarrierGLS, Status: domain.Status{Code: domain.StatusUnknown}}
	if err := uc.RefreshAll(ctx); err != nil {
		t.Fatal(err)
	}
	// A failing mailbox is reported in the settings.
	mail.err = errors.New("connection refused")
	clock.t = clock.t.Add(time.Hour)
	if err := uc.DiscoverParcels(ctx); err == nil {
		t.Fatal("expected an error")
	}
	if v, _ := uc.MailSettings(ctx); !strings.Contains(v.LastError, "connection refused") || v.LastRun.IsZero() {
		t.Fatalf("%+v", v)
	}
	// Another server without a new password is refused (#191).
	mail.err = nil
	if _, err := uc.SaveMailSettings(ctx, MailSettingsInput{Enabled: true, Host: "imap.evil.example", Username: "me", Folder: "Pakete"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("stored password sent to another server", err)
	}
}
