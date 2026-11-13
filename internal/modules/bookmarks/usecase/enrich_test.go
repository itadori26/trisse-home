package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/bookmarks/domain"
	"trisse-home/internal/shared/kernel"
)

type memExtras struct {
	settings domain.Settings
	links    map[string]domain.LinkStatus
	icons    map[string]domain.Icon
}

func newExtras() *memExtras {
	return &memExtras{links: map[string]domain.LinkStatus{}, icons: map[string]domain.Icon{}}
}

func (m *memExtras) LoadSettings(context.Context) (domain.Settings, error) { return m.settings, nil }
func (m *memExtras) SaveSettings(_ context.Context, s domain.Settings) error {
	m.settings = s
	return nil
}
func (m *memExtras) ListLinkStatuses(context.Context) (map[string]domain.LinkStatus, error) {
	out := map[string]domain.LinkStatus{}
	for k, v := range m.links {
		out[k] = v
	}
	return out, nil
}
func (m *memExtras) SaveLinkStatus(_ context.Context, id string, s domain.LinkStatus) error {
	m.links[id] = s
	return nil
}
func (m *memExtras) SaveIcon(_ context.Context, id string, icon domain.Icon) error {
	m.icons[id] = icon
	return nil
}
func (m *memExtras) Icon(_ context.Context, id string) (domain.Icon, error) {
	i, ok := m.icons[id]
	if !ok {
		return domain.Icon{}, domain.ErrIconNotFound
	}
	return i, nil
}
func (m *memExtras) IconIDs(context.Context) (map[string]bool, error) {
	out := map[string]bool{}
	for k := range m.icons {
		out[k] = true
	}
	return out, nil
}

type fakeWeb struct {
	pages    map[string]domain.PageMetadata
	codes    map[string]int
	fetched  []string
	checked  []string
	iconURLs []string
}

func (f *fakeWeb) Fetch(_ context.Context, u string) (domain.PageMetadata, error) {
	f.fetched = append(f.fetched, u)
	m, ok := f.pages[u]
	if !ok {
		return domain.PageMetadata{}, errors.New("unreachable")
	}
	return m, nil
}
func (f *fakeWeb) FetchIcon(_ context.Context, u string) (domain.Icon, error) {
	f.iconURLs = append(f.iconURLs, u)
	return domain.NewIcon("image/png", []byte{1, 2, 3})
}
func (f *fakeWeb) Check(_ context.Context, u string) (int, string, bool) {
	f.checked = append(f.checked, u)
	code, ok := f.codes[u]
	return code, u, !ok
}

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

func TestEnrichAndLinkChecks(t *testing.T) {
	ctx := context.Background()
	repo, extras := newMem(), newExtras()
	web := &fakeWeb{
		pages: map[string]domain.PageMetadata{"https://go.dev": {Title: " The Go\nLanguage ", IconURL: "/fav.png"}},
		codes: map[string]int{"https://go.dev": 200, "https://gone.example": 404},
	}
	c := &movingClock{t: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)}
	d := Deps{Repo: repo, Clock: c, IDs: &ids{}, Settings: extras, Links: extras, Icons: extras, Metadata: web, Checker: web}

	// Off by default: no outbound request at all.
	plain, err := NewCreateBookmark(d).Execute(ctx, Input{URL: "https://go.dev"})
	if err != nil || plain.Title != "go.dev" || len(web.fetched) != 0 {
		t.Fatalf("disabled enrichment fetched: %+v %v", web.fetched, err)
	}
	if n, _ := NewCheckLinks(d).Execute(ctx); n != 0 || len(web.checked) != 0 {
		t.Fatal("disabled link check ran")
	}
	if _, err := NewEnrichBookmark(d).Execute(ctx, plain.ID); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("enrich while disabled: %v", err)
	}
	if _, err := NewCheckLink(d).Execute(ctx, plain.ID); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("check while disabled: %v", err)
	}

	if _, err := NewSaveSettings(d).Execute(ctx, SettingsView{Enrich: true, CheckLinks: true}); err != nil {
		t.Fatal(err)
	}
	if s, _ := NewGetSettings(d).Execute(ctx); !s.Enrich || !s.CheckLinks {
		t.Fatalf("settings %+v", s)
	}

	// The title is taken from the page only when none was entered.
	v, err := NewCreateBookmark(d).Execute(ctx, Input{URL: "https://go.dev"})
	if err != nil || v.Title != "The Go Language" || !v.HasIcon || web.iconURLs[0] != "https://go.dev/fav.png" {
		t.Fatalf("enriched %+v %v %v", v, err, web.iconURLs)
	}
	own, _ := NewCreateBookmark(d).Execute(ctx, Input{URL: "https://go.dev", Title: "Mine"})
	if own.Title != "Mine" || !own.HasIcon {
		t.Fatalf("own title replaced: %+v", own)
	}
	// Unreadable pages keep the bookmark as entered.
	gone, err := NewCreateBookmark(d).Execute(ctx, Input{URL: "https://gone.example"})
	if err != nil || gone.Title != "gone.example" || gone.HasIcon {
		t.Fatalf("unreadable page: %+v %v", gone, err)
	}
	if _, err := NewEnrichBookmark(d).Execute(ctx, gone.ID); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatalf("explicit enrich of an unreadable page: %v", err)
	}
	// The explicit refresh fills the title of the bookmark created while disabled.
	if v, err := NewEnrichBookmark(d).Execute(ctx, plain.ID); err != nil || v.Title != "The Go Language" || !v.HasIcon {
		t.Fatalf("refresh %+v %v", v, err)
	}
	if icon, err := NewGetIcon(d).Execute(ctx, plain.ID); err != nil || icon.ContentType != "image/png" {
		t.Fatalf("icon %+v %v", icon, err)
	}

	// The job checks everything that was never checked, then nothing for a week.
	if n, err := NewCheckLinks(d).Execute(ctx); err != nil || n != 4 {
		t.Fatalf("checked %d %v", n, err)
	}
	list, _ := NewListBookmarks(d).Execute(ctx, Filter{Broken: true})
	if len(list.Bookmarks) != 1 || list.Bookmarks[0].ID != gone.ID || list.Bookmarks[0].Link.Code != 404 {
		t.Fatalf("broken filter %+v", list.Bookmarks)
	}
	if n, _ := NewCheckLinks(d).Execute(ctx); n != 0 {
		t.Fatalf("checked again within a week: %d", n)
	}
	c.t = c.t.Add(domain.LinkCheckInterval)
	web.codes["https://gone.example"] = 200
	if v, err := NewCheckLink(d).Execute(ctx, gone.ID); err != nil || v.Link.State != domain.LinkOK {
		t.Fatalf("manual check %+v %v", v, err)
	}
	if n, _ := NewCheckLinks(d).Execute(ctx); n != 3 {
		t.Fatalf("weekly recheck %d", n)
	}
}
