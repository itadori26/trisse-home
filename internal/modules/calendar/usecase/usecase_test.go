package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

type fakeSource struct{ fail map[string]bool }

var now = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func (f *fakeSource) Events(_ context.Context, s domain.Settings, u string, _, _ time.Time) ([]domain.Event, error) {
	if s.Password != "pw" {
		return nil, errors.New("401 Unauthorized")
	}
	if f.fail[u] {
		return nil, errors.New("404 Not Found")
	}
	name := domain.CalendarName(u)
	return []domain.Event{
		{Summary: name + " past", Start: now.Add(-3 * time.Hour), End: now.Add(-2 * time.Hour), Calendar: name},
		{Summary: name + " soon", Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour), Calendar: name},
		{Summary: name + " later", Start: now.Add(10 * 24 * time.Hour), End: now.Add(10*24*time.Hour + time.Hour), Calendar: name},
	}, nil
}

type clk struct{}

func (clk) Now() time.Time { return now }

func TestCalendar(t *testing.T) {
	ctx := context.Background()
	repo, src := &memRepo{}, &fakeSource{fail: map[string]bool{}}
	uc := NewCalendar(Deps{Repo: repo, Source: src, Writer: &fakeWriter{}, Subscriptions: &memSubs{}, Feeds: &fakeFeeds{}, IDs: &seqIDs{}, Clock: clk{}})
	if _, err := uc.Upcoming(ctx, 7); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("expected not configured, got %v", err)
	}
	urls := []string{"https://c.example/cal/home/", "https://c.example/cal/work/"}
	if _, err := uc.SaveSettings(ctx, SettingsInput{CalendarURLs: urls, Username: "me", Password: "bad"}); err == nil {
		t.Fatal("expected connection error")
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{CalendarURLs: urls, Username: "me", Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	week, _ := uc.Upcoming(ctx, 7)
	if len(week.Events) != 2 || !strings.HasSuffix(week.Events[0].Summary, "soon") {
		t.Fatalf("7-day agenda wrong: %+v", week.Events)
	}
	month, _ := uc.Upcoming(ctx, 31)
	if len(month.Events) != 4 {
		t.Fatalf("31-day agenda wrong: %d events", len(month.Events))
	}
	src.fail["https://c.example/cal/work/"] = true
	if err := uc.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	partial, _ := uc.Upcoming(ctx, 31)
	if len(partial.Events) != 2 || !strings.Contains(partial.Error, "work") {
		t.Fatalf("partial failure not reported: %+v", partial)
	}
}

type fakeWriter struct {
	created []domain.Event
	urls    []string
	err     error
}

func (f *fakeWriter) Create(_ context.Context, _ domain.Settings, u string, e domain.Event) error {
	f.created, f.urls = append(f.created, e), append(f.urls, u)
	return f.err
}

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return "id" + strings.Repeat("x", s.n) }

func TestRangeAndCreate(t *testing.T) {
	ctx := context.Background()
	repo, src, w := &memRepo{}, &fakeSource{fail: map[string]bool{}}, &fakeWriter{}
	uc := NewCalendar(Deps{Repo: repo, Source: src, Writer: w, Subscriptions: &memSubs{}, Feeds: &fakeFeeds{}, IDs: &seqIDs{}, Clock: clk{}})
	urls := []string{"https://c.example/cal/home/"}
	if _, err := uc.SaveSettings(ctx, SettingsInput{CalendarURLs: urls, Username: "me", Password: "pw"}); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Range(ctx, now, now.AddDate(0, 0, 60)); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("too long range accepted: %v", err)
	}
	inCache, err := uc.Range(ctx, now, now.AddDate(0, 0, 7))
	if err != nil || len(inCache.Events) == 0 {
		t.Fatalf("cached range: %+v %v", inCache, err)
	}
	past, err := uc.Range(ctx, now.AddDate(0, -2, 0), now.AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("live range: %v", err)
	}
	_ = past
	if _, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: "https://evil.example/", Summary: "x", Start: now, End: now.Add(time.Hour)}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown calendar accepted: %v", err)
	}
	day := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	v, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: urls[0], Summary: "Holiday", AllDay: true, Start: day, End: day})
	if err != nil || !v.End.Equal(day.AddDate(0, 0, 1)) || !strings.HasSuffix(w.created[0].UID, "@trisse-home") || v.Calendar != "home" {
		t.Fatalf("all-day create: %+v %v %+v", v, err, w.created)
	}
	w.err = errors.New("403 Forbidden")
	if _, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: urls[0], Summary: "x", Start: now, End: now.Add(time.Hour)}); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("server error not reported: %v", err)
	}
}

type memSubs struct{ list []domain.Subscription }

func (m *memSubs) List(context.Context) ([]domain.Subscription, error) {
	return append([]domain.Subscription(nil), m.list...), nil
}
func (m *memSubs) Get(_ context.Context, id string) (domain.Subscription, error) {
	for _, s := range m.list {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Subscription{}, domain.ErrSubscriptionNotFound
}
func (m *memSubs) Save(_ context.Context, s domain.Subscription) error {
	for i := range m.list {
		if m.list[i].ID == s.ID {
			m.list[i] = s
			return nil
		}
	}
	m.list = append(m.list, s)
	return nil
}
func (m *memSubs) Delete(_ context.Context, id string) error {
	for i := range m.list {
		if m.list[i].ID == id {
			m.list = append(m.list[:i], m.list[i+1:]...)
			return nil
		}
	}
	return nil
}

type fakeFeeds struct {
	fail  map[string]bool
	calls int
}

func (f *fakeFeeds) Events(_ context.Context, s domain.Subscription, from, to time.Time) ([]domain.Event, error) {
	f.calls++
	if f.fail[s.URL] {
		return nil, errors.New("404 Not Found")
	}
	return []domain.Event{{UID: s.ID + "-1", Summary: "Lecture", Start: now.Add(5 * time.Hour), End: now.Add(6 * time.Hour)}}, nil
}

func TestSubscriptions(t *testing.T) {
	ctx := context.Background()
	subs, feeds := &memSubs{}, &fakeFeeds{fail: map[string]bool{"https://bad.example/x.ics": true}}
	uc := NewCalendar(Deps{Repo: &memRepo{}, Source: &fakeSource{}, Writer: &fakeWriter{}, Subscriptions: subs, Feeds: feeds, IDs: &seqIDs{}, Clock: clk{}})

	if _, err := uc.AddSubscription(ctx, SubscriptionInput{Name: "Uni", URL: "https://bad.example/x.ics"}); !errors.Is(err, kernel.ErrInvalid) || len(subs.list) != 0 {
		t.Fatalf("a feed that cannot be loaded is not saved: %v", err)
	}
	v, err := uc.AddSubscription(ctx, SubscriptionInput{Name: "Uni", URL: "webcal://uni.example/plan.ics?token=secret", Color: "#AA0000", RefreshMinutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	if v.URLHint != "https://uni.example/…" || v.Color != "#aa0000" || v.RefreshMinutes != 60 || strings.Contains(v.URLHint, "secret") {
		t.Fatalf("view hides the token: %+v", v)
	}

	// Subscriptions alone make the calendar usable, without CalDAV.
	agenda, err := uc.Upcoming(ctx, 7)
	if err != nil || len(agenda.Events) != 1 || agenda.Events[0].Calendar != "Uni" || agenda.Events[0].Color != "#aa0000" {
		t.Fatalf("agenda from a subscription: %+v %v", agenda, err)
	}
	if month, err := uc.Range(ctx, now, now.Add(24*time.Hour)); err != nil || len(month.Events) != 1 {
		t.Fatalf("month view includes subscriptions: %+v %v", month, err)
	}

	calls := feeds.calls
	u, err := uc.UpdateSubscription(ctx, v.ID, SubscriptionInput{Name: "University", Color: "#00aa00", RefreshMinutes: 15})
	if err != nil || u.Name != "University" || subs.list[0].URL != "https://uni.example/plan.ics?token=secret" {
		t.Fatalf("an empty URL keeps the stored one: %+v %v", u, err)
	}
	if feeds.calls != calls+1 {
		t.Fatalf("an unchanged URL is not tested again, only refreshed: %d calls", feeds.calls-calls)
	}
	if _, err := uc.UpdateSubscription(ctx, v.ID, SubscriptionInput{Name: "x", RefreshMinutes: 5}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("interval below 15 minutes: %v", err)
	}
	if _, err := uc.UpdateSubscription(ctx, "nope", SubscriptionInput{Name: "x"}); !errors.Is(err, domain.ErrSubscriptionNotFound) {
		t.Fatalf("unknown: %v", err)
	}
	list, _ := uc.ListSubscriptions(ctx)
	if len(list) != 1 || list[0].Name != "University" {
		t.Fatalf("list: %+v", list)
	}

	if err := uc.DeleteSubscription(ctx, v.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Upcoming(ctx, 7); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("without any source the calendar is unconfigured again: %v", err)
	}
}

func TestSubscriptionLimit(t *testing.T) {
	ctx := context.Background()
	subs := &memSubs{}
	for i := range domain.MaxSubscriptions {
		subs.list = append(subs.list, domain.Subscription{ID: string(rune('a' + i)), Name: "x", URL: "https://x/", Color: domain.DefaultColor, RefreshInterval: time.Hour})
	}
	uc := NewCalendar(Deps{Repo: &memRepo{}, Source: &fakeSource{}, Writer: &fakeWriter{}, Subscriptions: subs, Feeds: &fakeFeeds{}, IDs: &seqIDs{}, Clock: clk{}})
	if _, err := uc.AddSubscription(ctx, SubscriptionInput{Name: "one more", URL: "https://y/"}); !errors.Is(err, domain.ErrTooManySubscriptions) {
		t.Fatalf("limit: %v", err)
	}
}

func TestCalDAVAndSubscriptionMerged(t *testing.T) {
	ctx := context.Background()
	subs := &memSubs{list: []domain.Subscription{{ID: "s", Name: "Holidays", URL: "https://bad.example/x.ics", Color: "#00aa00", RefreshInterval: time.Hour}}}
	repo := &memRepo{s: &domain.Settings{CalendarURLs: []string{"https://c.example/cal/home/"}, Username: "me", Password: "pw"}}
	uc := NewCalendar(Deps{Repo: repo, Source: &fakeSource{fail: map[string]bool{}}, Writer: &fakeWriter{}, Subscriptions: subs,
		Feeds: &fakeFeeds{fail: map[string]bool{"https://bad.example/x.ics": true}}, IDs: &seqIDs{}, Clock: clk{}})
	agenda, err := uc.Upcoming(ctx, 7)
	if err != nil || len(agenda.Events) != 1 || !strings.Contains(agenda.Error, "Holidays") {
		t.Fatalf("a failing subscription is reported, CalDAV events still shown: %+v %v", agenda, err)
	}
}
