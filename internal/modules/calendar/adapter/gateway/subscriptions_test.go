package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/platform/database/dbtest"
)

type upperCipher struct{}

func (upperCipher) Encrypt(s string) (string, error) { return "enc:" + strings.ToUpper(s), nil }
func (upperCipher) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "enc:") {
		return "", errors.New("bad ciphertext")
	}
	return strings.ToLower(strings.TrimPrefix(s, "enc:")), nil
}

func TestSubscriptionRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "calendar", Migrations())
	repo := NewSubscriptionRepository(db, upperCipher{})
	a, _ := domain.NewSubscription("a", "Uni", "https://uni.example/plan.ics?token=secret", "#aa0000", time.Hour)
	b, _ := domain.NewSubscription("b", "Holidays", "https://h.example/de.ics", "", 0)
	for _, s := range []domain.Subscription{a, b} {
		if err := repo.Save(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT url_encrypted FROM calendar_subscriptions WHERE id = 'a'`).Scan(&stored)
	if strings.Contains(stored, "secret") {
		t.Fatalf("URL stored in plain text: %q", stored)
	}
	a.Name = "University"
	if err := repo.Save(ctx, a); err != nil {
		t.Fatal(err)
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 || list[0] != a || list[1] != b {
		t.Fatalf("list keeps creation order after an update: %+v %v", list, err)
	}
	if err := repo.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, "a"); !errors.Is(err, domain.ErrSubscriptionNotFound) {
		t.Fatalf("deleted: %v", err)
	}
}

const weekly = "BEGIN:VCALENDAR\r\nVERSION:2.0\r\n" +
	"BEGIN:VEVENT\r\nUID:lecture\r\nSUMMARY:Lecture\r\nDTSTART:20261102T080000Z\r\nDTEND:20261102T093000Z\r\nRRULE:FREQ=WEEKLY;COUNT=10\r\nEND:VEVENT\r\n" +
	"BEGIN:VEVENT\r\nUID:day\r\nSUMMARY:Holiday\r\nDTSTART;VALUE=DATE:20261225\r\nEND:VEVENT\r\n" +
	"END:VCALENDAR\r\n"

func TestFeed(t *testing.T) {
	var downloads, notModified atomic.Int32
	fail := atomic.Bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		if r.Header.Get("If-None-Match") == `"v1"` {
			notModified.Add(1)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		downloads.Add(1)
		w.Header().Set("ETag", `"v1"`)
		_, _ = w.Write([]byte(weekly))
	}))
	defer srv.Close()

	now := time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)
	feed := NewFeed(srv.Client(), time.UTC, func() time.Time { return now })
	sub := domain.Subscription{ID: "s", Name: "Uni", URL: srv.URL + "/plan.ics", RefreshInterval: time.Hour}
	ctx := context.Background()

	events, err := feed.Events(ctx, sub, now, now.Add(45*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	// Six weekly lectures from 16 Nov to 21 Dec, then the holiday on 25 Dec.
	if len(events) != 7 || events[0].Start != time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC) || !events[6].AllDay {
		t.Fatalf("weekly rule expanded within the range plus the holiday: %+v", events)
	}

	// Within the interval the cached copy is used; afterwards a conditional request.
	_, _ = feed.Events(ctx, sub, now, now.Add(24*time.Hour))
	if downloads.Load() != 1 || notModified.Load() != 0 {
		t.Fatalf("cached within the interval: %d downloads, %d 304", downloads.Load(), notModified.Load())
	}
	now = now.Add(time.Hour)
	if _, err := feed.Events(ctx, sub, now, now.Add(24*time.Hour)); err != nil || notModified.Load() != 1 || downloads.Load() != 1 {
		t.Fatalf("conditional request after the interval: %v, %d downloads, %d 304", err, downloads.Load(), notModified.Load())
	}

	// A failing server keeps the last good copy.
	fail.Store(true)
	now = now.Add(time.Hour)
	if events, err := feed.Events(ctx, sub, now, now.Add(31*24*time.Hour)); err != nil || len(events) == 0 {
		t.Fatalf("stale copy on errors: %v", err)
	}
	other := sub
	other.ID = "new"
	if _, err := feed.Events(ctx, other, now, now.Add(time.Hour)); err == nil {
		t.Fatal("a feed that was never loaded reports the error")
	}
}
