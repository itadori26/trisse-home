package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	m        map[string]domain.Article
	images   map[string]map[int]Image
	extra    int64
	settings domain.Settings
}

func (r *memRepo) Pending(_ context.Context, limit int) ([]domain.Article, error) {
	var out []domain.Article
	for _, a := range r.m {
		if a.Pending && len(out) < limit {
			out = append(out, a)
		}
	}
	return out, nil
}
func (r *memRepo) Settings(context.Context) (domain.Settings, error) { return r.settings, nil }
func (r *memRepo) SaveSettings(_ context.Context, s domain.Settings) error {
	r.settings = s
	return nil
}

func newRepo() *memRepo {
	return &memRepo{m: map[string]domain.Article{}, images: map[string]map[int]Image{}}
}
func (r *memRepo) List(context.Context) ([]domain.Article, error) {
	var out []domain.Article
	for _, a := range r.m {
		out = append(out, a)
	}
	return out, nil
}
func (r *memRepo) Get(_ context.Context, id string) (domain.Article, error) {
	a, ok := r.m[id]
	if !ok {
		return a, domain.ErrNotFound
	}
	return a, nil
}
func (r *memRepo) Save(_ context.Context, a domain.Article, images map[int]Image) error {
	r.m[a.ID], r.images[a.ID] = a, images
	return nil
}
func (r *memRepo) SetStatus(_ context.Context, id string, s domain.Status, at time.Time) error {
	a := r.m[id]
	a.Status, a.StatusAt = s, at
	r.m[id] = a
	return nil
}
func (r *memRepo) Delete(_ context.Context, id string) error {
	delete(r.m, id)
	delete(r.images, id)
	return nil
}
func (r *memRepo) Image(_ context.Context, id string, n int) (Image, error) {
	img, ok := r.images[id][n]
	if !ok {
		return Image{}, domain.ErrNotFound
	}
	return img, nil
}
func (r *memRepo) TotalBytes(context.Context) (int64, error) {
	t := r.extra
	for _, a := range r.m {
		t += a.Bytes
	}
	return t, nil
}

type fakePages struct {
	page Page
	err  error
}

func (f *fakePages) Read(context.Context, string) (Page, error) { return f.page, f.err }

type fakeImages struct {
	size  int
	fails map[string]bool
	calls int
}

func (f *fakeImages) Fetch(_ context.Context, url string) (Image, error) {
	f.calls++
	if f.fails[url] {
		return Image{}, errors.New("404")
	}
	return Image{Bytes: make([]byte, f.size), ContentType: "image/png"}, nil
}

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { c.t = c.t.Add(time.Minute); return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return "a" + strconv.Itoa(i.n) }

func TestSave(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	pages := &fakePages{page: Page{Title: "Post", Site: "Blog", Content: `<p>x</p><img src="cid:0" alt="a"><img src="cid:1" alt="b">`,
		Images: []string{"https://b.example/0.png", "https://b.example/1.png"}, Words: 500}}
	imgs := &fakeImages{size: 1000, fails: map[string]bool{"https://b.example/1.png": true}}
	uc := New(repo, pages, imgs, codec{}, &clk{t: time.Unix(0, 0)}, &ids{})

	a, err := uc.Save(ctx, "https://b.example/post#top")
	if err != nil || a.Title != "Post" || a.ReadingMinutes() != 3 || a.Error != "" {
		t.Fatalf("%+v %v", a, err)
	}
	if a.Content != `<p>x</p><img src="/api/readlater/a1/images/0" alt="a">` || a.Bytes != int64(len(a.Content))+1000 {
		t.Fatalf("images not resolved: %q %d", a.Content, a.Bytes)
	}
	if img, err := uc.Image(ctx, "a1", 0); err != nil || len(img.Bytes) != 1000 {
		t.Fatalf("image: %v", err)
	}
	// Saving the same page again returns the existing article.
	if again, _ := uc.Save(ctx, "https://b.example/post"); again.ID != "a1" || len(repo.m) != 1 {
		t.Fatalf("duplicate: %+v", again)
	}

	// An unreachable page keeps the link with the reason.
	pages.err = kernel.NewPublicError(kernel.ErrInvalid, "The server answered 404 Not Found.")
	b, err := uc.Save(ctx, "https://b.example/gone")
	if err != nil || b.Content != "" || !strings.Contains(b.Error, "404") {
		t.Fatalf("%+v %v", b, err)
	}
}

func TestLimits(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	var content strings.Builder
	var srcs []string
	for i := range 30 {
		content.WriteString(`<img src="cid:` + strconv.Itoa(i) + `" alt="">`)
		srcs = append(srcs, "https://b.example/"+strconv.Itoa(i))
	}
	imgs := &fakeImages{size: domain.MaxImageBytes}
	uc := New(repo, &fakePages{page: Page{Content: content.String(), Images: srcs}}, imgs, codec{}, &clk{}, &ids{})
	a, err := uc.Save(ctx, "https://b.example/many")
	if err != nil {
		t.Fatal(err)
	}
	// 8 MB per article: four 2 MB images.
	if got := strings.Count(a.Content, "<img"); got != 4 || len(repo.images[a.ID]) != 4 {
		t.Fatalf("images kept: %d", got)
	}
	repo.extra = domain.MaxStorageBytes
	if _, err := uc.Save(ctx, "https://b.example/full"); !errors.Is(err, kernel.ErrConflict) {
		t.Fatalf("storage full: %v", err)
	}
}

func TestStatesAndList(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	uc := New(repo, &fakePages{page: Page{Content: "<p>x</p>"}}, &fakeImages{}, codec{}, &clk{t: time.Unix(0, 0)}, &ids{})
	a, _ := uc.Save(ctx, "https://x.example/1")
	b, _ := uc.Save(ctx, "https://x.example/2")
	if err := uc.SetStatus(ctx, a.ID, domain.StatusArchived); err != nil {
		t.Fatal(err)
	}
	if err := uc.SetStatus(ctx, a.ID, "gone"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
	if list, _ := uc.List(ctx, ""); len(list) != 1 || list[0].ID != b.ID || list[0].Content != "" {
		t.Fatalf("not archived: %+v", list)
	}
	if list, _ := uc.List(ctx, domain.StatusArchived); len(list) != 1 || list[0].ID != a.ID {
		t.Fatalf("archived: %+v", list)
	}
	if err := uc.Delete(ctx, a.ID); err != nil || len(repo.m) != 1 {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, a.ID); !errors.Is(err, kernel.ErrNotFound) {
		t.Fatal(err)
	}
}

type codec struct{ entries []domain.ImportEntry }

func (c codec) Pocket([]byte) ([]domain.ImportEntry, error) {
	if c.entries == nil {
		return nil, errors.New("not csv")
	}
	return c.entries, nil
}

func TestImportFetchAndCleanup(t *testing.T) {
	ctx := context.Background()
	repo := newRepo()
	pages := &fakePages{page: Page{Title: "Echt", Content: "<p>Text</p>", Words: 400}}
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	c := &clk{t: now}
	existing, _ := domain.NewArticle("x", "https://a.example/known", now)
	repo.m["x"] = existing
	entries := []domain.ImportEntry{
		{URL: "https://a.example/1", Title: "Eins", Added: now.AddDate(-1, 0, 0)},
		{URL: "https://a.example/2", Title: "Zwei", Archived: true},
		{URL: "https://a.example/known"},
		{URL: "javascript:alert(1)"},
	}
	for i := range 6 {
		entries = append(entries, domain.ImportEntry{URL: "https://b.example/" + strconv.Itoa(i)})
	}
	uc := New(repo, pages, &fakeImages{}, codec{entries: entries}, c, &ids{})
	r, err := uc.ImportPocket(ctx, []byte("csv"))
	if err != nil || r.Added != 8 || r.Duplicates != 1 || r.Invalid != 1 {
		t.Fatalf("%+v %v", r, err)
	}
	var one domain.Article
	for _, a := range repo.m {
		if a.URL == "https://a.example/1" {
			one = a
		}
	}
	if !one.Pending || one.Title != "Eins" || !one.SavedAt.Equal(now.AddDate(-1, 0, 0)) || one.Content != "" {
		t.Fatalf("imported: %+v", one)
	}
	if _, err := New(repo, pages, &fakeImages{}, codec{}, c, &ids{}).ImportPocket(ctx, []byte("x")); err == nil {
		t.Fatal("not an export")
	}
	// The job makes a few reader copies per run.
	_ = uc.FetchPending(ctx)
	left, _ := repo.Pending(ctx, 100)
	if len(left) != 8-PendingPerRun {
		t.Fatalf("pending after one run: %d", len(left))
	}
	for range 3 {
		_ = uc.FetchPending(ctx)
	}
	if left, _ := repo.Pending(ctx, 100); len(left) != 0 {
		t.Fatal("all copies made")
	}
	if got := repo.m[one.ID]; got.Content == "" || got.Pending || got.Title != "Echt" {
		t.Fatalf("filled: %+v", got)
	}
	// Cleanup: off by default, then archived articles older than the setting go.
	if _, err := uc.SaveSettings(ctx, 400); err == nil {
		t.Fatal("days bounded")
	}
	_ = uc.Cleanup(ctx)
	n := len(repo.m)
	_, _ = uc.SaveSettings(ctx, 30)
	_ = uc.SetStatus(ctx, one.ID, domain.StatusArchived)
	c.t = now.AddDate(0, 0, 31)
	_ = uc.Cleanup(ctx)
	if _, ok := repo.m[one.ID]; ok || len(repo.m) != n-2 {
		t.Fatalf("cleanup: %d of %d left", len(repo.m), n)
	}
}
