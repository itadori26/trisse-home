package domain

import (
	"errors"
	"testing"
	"time"

	"trisse-home/internal/shared/kernel"
)

func TestCheckDigits(t *testing.T) {
	if !gs1Check("4006381333931") || gs1Check("4006381333932") {
		t.Fatal("GS1 mod 10")
	}
	if !upsCheck("1Z999AA10123456784") || upsCheck("1Z999AA10123456785") {
		t.Fatal("UPS check digit")
	}
}

// Anonymised fixtures of real shipping mails (German and English).
func TestExtractCandidates(t *testing.T) {
	for name, tc := range map[string]struct {
		mail Mail
		want []Candidate
	}{
		"dhl announcement": {
			Mail{From: "DHL Paket <noreply@dhl.de>", Subject: "Ihre DHL Sendung von Zalando kommt heute",
				Text: "Hallo,\nIhre Sendung ist unterwegs.\nSendungsnummer: 00340434161094042557\nAuf dhl.de verfolgen: https://www.dhl.de/de/privatkunden/pakete-empfangen/verfolgen.html?piececode=00340434161094042557\nBei Fragen: 0228 4333112"},
			[]Candidate{{Number: "00340434161094042557", Carrier: CarrierDHL, Label: "Zalando kommt heute"}},
		},
		"shop with dhl link": {
			Mail{From: "Thalia <service@thalia.de>", Subject: "Deine Bestellung 123456789 wurde versendet",
				Text: "Bestellnummer: 123456789012\nDein Paket wird von DHL geliefert.\nSendungsnummer 1234 5678 9012\nKundenservice: +49 251 5309 3000"},
			[]Candidate{{Number: "123456789012", Carrier: CarrierDHL, Label: "Thalia"}},
		},
		"hermes": {
			Mail{From: "\"OTTO\" <versand@otto.de>", Subject: "Ihre Lieferung ist unterwegs",
				Text: "Versand mit Hermes. Ihre Sendungsnummer lautet: 01234567890123. Sendungsverfolgung: https://www.myhermes.de/empfangen/sendungsverfolgung/sendungsinformation#01234567890123"},
			[]Candidate{{Number: "01234567890123", Carrier: CarrierHermes, Label: "OTTO"}},
		},
		"dpd link only": {
			Mail{From: "Shop <a@shop.example>", Subject: "Shipped",
				Text: `<a href="https://tracking.dpd.de/status/de_DE/parcel/01505012345678">Track</a>`},
			[]Candidate{{Number: "01505012345678", Carrier: CarrierDPD, Label: "Shop"}},
		},
		"gls": {
			Mail{From: "GLS <noreply@gls-pakete.de>", Subject: "Your parcel from Kaufland is on its way",
				Text: "Parcel number: 12345678901\nOrder number: 98765432109"},
			[]Candidate{{Number: "12345678901", Carrier: CarrierGLS, Label: "Kaufland is on its way"}},
		},
		"ups anywhere": {
			Mail{From: "Amazon.de <versandbestaetigung@amazon.de>", Subject: "Versendet: \"Kabel\"",
				Text: "Ihr Paket wird mit UPS zugestellt: 1Z999AA10123456784"},
			[]Candidate{{Number: "1Z999AA10123456784", Carrier: CarrierUPS, Label: "Amazon.de"}},
		},
		"order and phone numbers do not match": {
			Mail{From: "Shop <a@shop.example>", Subject: "Bestellbestätigung 4711",
				Text: "Bestellnummer: 304-1234567-1234567\nRechnungsnummer 2026112345\nTelefon 030 12345678\nIBAN DE89370400440532013000"},
			nil,
		},
		"unknown carrier is skipped": {
			Mail{From: "Shop <a@shop.example>", Subject: "Versand", Text: "Sendungsnummer: 987654321098"},
			nil,
		},
		"wrong check digit is skipped": {
			Mail{From: "DHL <noreply@dhl.de>", Subject: "Sendung", Text: "Sendungsnummer: 00340434161094042558"},
			nil,
		},
	} {
		got := ExtractCandidates(tc.mail)
		if len(got) != len(tc.want) {
			t.Errorf("%s: got %+v, want %+v", name, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("%s: got %+v, want %+v", name, got[i], tc.want[i])
			}
		}
	}
}

func TestMailSettings(t *testing.T) {
	s, err := NewMailSettings(true, " IMAP.example.com ", 0, "me@example.com", "app pw \"x\"", "", 0)
	if err != nil || s.Host != "imap.example.com" || s.Port != 993 || s.Folder != "INBOX" || s.Interval != DefaultMailInterval {
		t.Fatalf("%+v %v", s, err)
	}
	if off, err := NewMailSettings(false, "", 0, "", "", "", 0); err != nil || off.Enabled {
		t.Fatal("disabled without data is fine", err)
	}
	for name, err := range map[string]error{
		"host":     func() error { _, e := NewMailSettings(true, "imap.example.com/x", 993, "u", "p", "", 0); return e }(),
		"port":     func() error { _, e := NewMailSettings(true, "imap.example.com", 70000, "u", "p", "", 0); return e }(),
		"password": func() error { _, e := NewMailSettings(true, "imap.example.com", 993, "u", "", "", 0); return e }(),
		"newline": func() error {
			_, e := NewMailSettings(true, "imap.example.com", 993, "u", "p", "Pakete\r\nA1 DELETE x", 0)
			return e
		}(),
		"interval": func() error {
			_, e := NewMailSettings(true, "imap.example.com", 993, "u", "p", "", time.Minute)
			return e
		}(),
	} {
		if !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestTrackingURL(t *testing.T) {
	p := Parcel{Carrier: CarrierGLS, TrackingNumber: "12345678901"}
	if p.TrackingURL() != "https://www.gls-pakete.de/sendungsverfolgung?match=12345678901" {
		t.Fatal(p.TrackingURL())
	}
}
